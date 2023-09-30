package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/uptrace/bunrouter"
)

var tplIndex = pongo2.Must(pongo2.FromFile("templates/index.html"))
var tplGallery = pongo2.Must(pongo2.FromFile("templates/gallery.html"))

func webServer() {
	http.DefaultClient.Timeout = 120 * time.Second

	router := bunrouter.New()
	router.GET("/", webRoot)
	router.GET("/gallery", webGallery)
	router.GET("/grading", webGrading)
	router.GET("/api/history", webAPIHistory)
	router.POST("/api/genimg", webAPIGenImg)
	router.POST("/api/grade", webGrade)

	fs := http.FileServer(http.Dir("./media"))
	router.GET("/media/*path", bunrouter.HTTPHandler(http.StripPrefix("/media/", fs)))

	log.Println("Listening at", os.Getenv("LISTEN_ADDRESS"))
	err := http.ListenAndServe(os.Getenv("LISTEN_ADDRESS"), router)
	if err != nil {
		panic(err)
	}
}

func webRoot(w http.ResponseWriter, r bunrouter.Request) (err error) {
	tplCtx := pongo2.Context{}
	return tplIndex.ExecuteWriter(tplCtx, w)
}

func webGallery(w http.ResponseWriter, r bunrouter.Request) (err error) {
	tplCtx := pongo2.Context{}
	return tplGallery.ExecuteWriter(tplCtx, w)
}

func webAPIHistory(w http.ResponseWriter, r bunrouter.Request) (err error) {
	maxReturned := 100
	maxString := r.URL.Query().Get("max")
	if maxString != "" {
		max, err := strconv.Atoi(maxString)
		if err == nil {
			maxReturned = max
		} else {
			log.Println("error in max param:", err)
		}
	}
	ctx := r.Context()
	var history []DbHistory
	err = db.NewSelect().
		Model(&history).
		Order("ts DESC").
		Limit(maxReturned).
		Scan(ctx)
	if err != nil {
		log.Println(err)
		return
	}

	prunedHistory := make([]DbHistory, 0, len(history))
	for _, h := range history {
		fileName := fmt.Sprintf("media/%s", h.ImageFilename)
		if fileExists(fileName) && !isHtml(fileName) {
			prunedHistory = append(prunedHistory, h)
		}
	}

	return bunrouter.JSON(w, WebResponseHistory{
		Ok:             true,
		MediaURLPrefix: os.Getenv("MEDIA_URL"),
		History:        prunedHistory,
	})
}

func webAPIGenImg(w http.ResponseWriter, r bunrouter.Request) (err error) {
	ctx := r.Context()

	var req WebGenImgRequest
	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println(err)
		return
	}

	resolution, err := strconv.Atoi(os.Getenv("SD_RESOLUTION"))
	if err != nil {
		log.Println(err)
		return
	}

	sdReq := SDAPIRequest{
		Key:               os.Getenv("SD_API_KEY"),
		ModelID:           req.Model,
		Prompt:            strings.TrimSpace(req.Prompt),
		NegativePrompt:    "extra fingers, mutated hands, poorly drawn hands, poorly drawn face, deformed, ugly, blurry, bad anatomy, bad proportions, extra limbs, cloned face, skinny, glitchy, double torso, extra arms, extra hands, mangled fingers, missing lips, ugly face, distorted face, extra legs",
		Samples:           1,
		Width:             resolution,
		Height:            resolution,
		NumInferenceSteps: 41,
		GuidanceScale:     13.5,
		MultiLingual:      "yes",
		EnhancePrompt:     "no",
		SelfAttention:     "yes",
	}
	sdBody, err := json.Marshal(sdReq)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf(`User: "%s", Model: "%s", Prompt: "%s"`, req.Username, req.Model, req.Prompt)
	t0 := time.Now()

	sdHttpReq, err := http.NewRequest("POST", os.Getenv("SD_API_URL"), bytes.NewBuffer(sdBody))
	if err != nil {
		log.Println(err)
		return
	}
	sdHttpReq.Header.Add("Content-type", "application/json")
	sdHttpRes, err := http.DefaultClient.Do(sdHttpReq)
	if err != nil {
		log.Println(err)
		return
	}
	if sdHttpRes.StatusCode != http.StatusOK {
		log.Println(sdHttpRes.Status)
	}

	defer sdHttpRes.Body.Close()
	sdRespBody, err := ioutil.ReadAll(sdHttpRes.Body)
	if err != nil {
		log.Println(err)
		return
	}
	var sdResp SDAPIResult
	err = json.Unmarshal(sdRespBody, &sdResp)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Elapsed time:", time.Since(t0))

	if sdResp.Status == "processing" {
		log.Println("Waiting some more for image processing...")
		if len(sdResp.FutureLinks) == 0 {
			log.Println("API returned no output?", string(sdRespBody))
			return bunrouter.JSON(w, WebGenImgResponse{
				Ok: false,
			})
		}
		url := sdResp.FutureLinks[0]
		waitForHttpOkContent(url, "image/png")
		sdResp.Output = sdResp.FutureLinks
	} else {
		if len(sdResp.Output) == 0 {
			log.Println("API returned no output?", string(sdRespBody))
			return bunrouter.JSON(w, WebGenImgResponse{
				Ok: false,
			})
		}
	}

	urlParts := strings.Split(sdResp.Output[0], "/")
	imageFilename := urlParts[len(urlParts)-1]
	imagePath := fmt.Sprintf("media/%s", imageFilename)
	imageURL := fmt.Sprintf("%s/%s", os.Getenv("MEDIA_URL"), imageFilename)

	err = downloadFile(sdResp.Output[0], imagePath)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = db.NewInsert().
		Model(&DbHistory{
			Timestamp:     t0,
			Username:      req.Username,
			Model:         req.Model,
			Prompt:        req.Prompt,
			ImageFilename: imageFilename,
			Result:        sdResp,
		}).
		Exec(ctx)
	if err != nil {
		log.Println(err)
		return
	}

	return bunrouter.JSON(w, WebGenImgResponse{
		Ok:       true,
		Result:   sdResp,
		ImageURL: imageURL,
	})
}

func isHtml(fileName string) bool {
	f, _ := os.Open(fileName)
	defer f.Close()
	b := make([]byte, 1)
	f.Read(b)
	return b[0] == '<'
}

func waitForHttpOkContent(url, mime string) (ok bool, err error) {
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return
	}
	for {
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return false, err
		}
		log.Println("Status code:", res.StatusCode)
		if res.StatusCode == http.StatusOK && res.Header.Get("Content-type") == mime {
			return true, nil
		}
		log.Println("Still waiting for", url)
		time.Sleep(1 * time.Second)
	}
}
