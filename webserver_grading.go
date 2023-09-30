package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/flosch/pongo2/v6"
	"github.com/uptrace/bunrouter"
)

var tplGrading = pongo2.Must(pongo2.FromFile("templates/grading.html"))

func webGrading(w http.ResponseWriter, r bunrouter.Request) (err error) {
	tplCtx := pongo2.Context{}
	return tplGrading.ExecuteWriter(tplCtx, w)
}

func webGrade(w http.ResponseWriter, r bunrouter.Request) (err error) {
	var req WebGradeRequest
	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = db.NewInsert().
		Model(&DbGrades{
			WebGradeRequest: req,
		}).
		On("CONFLICT(grader_name, proposal_id) DO UPDATE").
		Set("grade = EXCLUDED.grade").
		Exec(r.Context())

	if err != nil {
		log.Println(err)
		return err
	}

	return bunrouter.JSON(w, WebGradeResponse{
		Ok: true,
	})
}
