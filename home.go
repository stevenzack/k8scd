package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/stevenzack/k8scd/store"
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
		v, e := store.ValidateRepoForm(r)
		if e != nil {
			badRequest(w, e)
			return
		}
		v.Id = newID()
		v.UpdatedAt = time.Now().Format(time.DateTime)
		v.CreatedAt = time.Now().Format(time.DateTime)

		// test
		dir := v.Id
		defer os.RemoveAll(dir)
		e = v.CloneGitRepoTo(dir)
		if e != nil {
			log.Println(e)
			badRequest(w, e)
			return
		}

		dst := filepath.Join(dir, v.YamlRelPath)
		b, e := os.ReadFile(dst)
		if e != nil {
			log.Println(e)
			badRequest(w, e)
			return
		}
		s := string(b)
		if !strings.Contains(s, tagReplacementIdentifier) {
			e = fmt.Errorf("invalid YAML file for k8s, tag identifier %s not exists at file: %s", tagReplacementIdentifier, v.YamlRelPath)
			log.Println(e)
			badRequest(w, e)
			return
		}

		e = stores.InsertRepo(v)
		if e != nil {
			log.Panic(e)
			return
		}

	case http.MethodPatch:
		v, e := store.ValidateRepoForm(r)
		if e != nil {
			badRequest(w, e)
			return
		}
		v.Id = r.FormValue("id")
		vs, e := stores.GetRepos()
		if e != nil {
			log.Panic(e)
			return
		}
		for _, v2 := range vs {
			if v2.Id == v.Id {
				e = stores.UpdateRepo(v)
				if e != nil {
					log.Panic(e)
					return
				}
				return
			}
		}
		badRequest(w, "item with id "+v.Id+" not found")
		return
	case http.MethodDelete:
		id := r.FormValue("id")
		vs, e := stores.GetRepos()
		if e != nil {
			log.Panic(e)
			return
		}
		for _, v2 := range vs {
			if v2.Id == id {
				e = stores.DeleteRepo(id)
				if e != nil {
					log.Panic(e)
					return
				}
				return
			}
		}
		badRequest(w, "item not found")
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}
