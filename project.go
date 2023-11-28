package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func project(w http.ResponseWriter, r *http.Request) {
	if auth(w, r) {
		return
	}
	projectId := r.URL.Path[len("/projects/"):]
	if strings.HasSuffix(projectId, "/deployments/") {
		// /projects/{id}/deployments/
		projectId = projectId[:len(projectId)-len("/deployments/")]
		switch r.Method {
		case http.MethodPost:
			tag := r.FormValue("tag")
			if tag == "" {
				badRequest(w, "tag cannot be empty")
				return
			}

			var req WebHookRequest
			req.PushData.Tag = tag
			b, e := json.Marshal(req)
			if e != nil {
				log.Panic(e)
				return
			}
			res, e := http.Post(r.Header.Get("Origin")+dockerHubWebhookRoutePath+projectId, "application/json", bytes.NewReader(b))
			if e != nil {
				log.Println(e)
				badRequest(w, e.Error())
				return
			}
			defer res.Body.Close()
			b, e = io.ReadAll(res.Body)
			if e != nil {
				log.Panic(e)
				return
			}
			if res.StatusCode != 200 {
				e = fmt.Errorf(string(b))
				log.Println(e)
				badRequest(w, e.Error())
				return
			}
		default:
			badRequest(w, "method not allowed")
		}
		return
	}
	if !validateID(projectId) {
		badRequest(w, "invalid project id")
		return
	}

	project, e := stores.GetRepoById(projectId)
	if e != nil {
		badRequest(w, e)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// pageStr := r.FormValue("page")
		// var page = 0
		// if pageStr != "" {
		// 	page, e = strconv.Atoi(pageStr)
		// 	if e != nil {
		// 		badRequest(w, "invalid page:"+e.Error())
		// 		return
		// 	}
		// }

		vs, e := stores.QueryDeployment(projectId)
		if e != nil {
			log.Panic(e)
			return
		}
		project.Deployments = vs
		t.ExecuteTemplate(w, "project.html", project)
		return
	}
}
