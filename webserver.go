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

func webServer() {
	router := bunrouter.New()
	router.GET("/", webRoot)
	router.GET("/api/history", webAPIHistory)
	router.POST("/api/genimg", webAPIGenImg)

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
	err = tplIndex.ExecuteWriter(tplCtx, w)
	if err != nil {
		log.Println(err)
	}
	return
}

func webAPIHistory(w http.ResponseWriter, r bunrouter.Request) (err error) {
	ctx := r.Context()
	var history []DbHistory
	err = db.NewSelect().
		Model(&history).
		Order("ts DESC").
		Limit(100).
		Scan(ctx)
	if err != nil {
		log.Println(err)
		return
	}
	return bunrouter.JSON(w, WebResponseHistory{
		Ok:             true,
		MediaURLPrefix: os.Getenv("MEDIA_URL"),
		History:        history,
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
		Prompt:            req.Prompt,
		NegativePrompt:    "extra fingers, mutated hands, poorly drawn hands, poorly drawn face, deformed, ugly, blurry, bad anatomy, bad proportions, extra limbs, cloned face, skinny, glitchy, double torso, extra arms, extra hands, mangled fingers, missing lips, ugly face, distorted face, extra legs",
		Samples:           1,
		Width:             resolution,
		Height:            resolution,
		NumInferenceSteps: 30,
		GuidanceScale:     7.5,
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

	if len(sdResp.Output) == 0 {
		log.Println("API returned no output?", string(sdRespBody))
		return bunrouter.JSON(w, WebGenImgResponse{
			Ok: false,
		})
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
