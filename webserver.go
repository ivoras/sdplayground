package main

import (
	"log"
	"net/http"
	"os"

	"github.com/flosch/pongo2/v6"
	"github.com/uptrace/bunrouter"
)

var tplIndex = pongo2.Must(pongo2.FromFile("templates/index.html"))

func webServer() {
	router := bunrouter.New()
	router.GET("/", webRoot)
	router.GET("/api/history", webAPIHistory)

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
		Ok:      true,
		History: history,
	})
}
