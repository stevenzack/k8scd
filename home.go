package main

import (
	"log"
	"net/http"
)

func home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if auth(w, r) {
		return
	}

	switch r.Method {
	case http.MethodGet:
		vs, e := stores.GetRepos()
		if e != nil {
			log.Panic(e)
			return
		}

		t.ExecuteTemplate(w, "index.html", vs)
		return
	case http.MethodPost:

	}
}
