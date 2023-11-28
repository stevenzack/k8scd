package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	apiNotifierRoutePath     = "/api/notifier/"
	tagReplacementIdentifier = "{{.Tag}}"
)

// tags: name/app:latest,name/app:1.0.0
func notifier(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) <= len(apiNotifierRoutePath) {
		return
	}
	id := r.URL.Path[len(apiNotifierRoutePath):]
	if id == "" || strings.Contains(id, "|") {
		badRequest(w, "project id empty or invalid")
		return
	}
	tag, e := parseTags(r.FormValue("tags"))
	if e != nil {
		badRequest(w, e.Error())
		return
	}

	project, e := stores.GetRepoById(id)
	if e != nil {
		badRequest(w, "project id empty or invalid:"+id)
		return
	}
	defer stores.UpdateRepo(project)

	//do
	dir := filepath.Join("repos-cache", project.Id)
	defer os.RemoveAll(dir)
	e = project.CloneGitRepoTo(dir)
	if e != nil {
		log.Panic(e)
		project.LastError = e.Error()
		return
	}

	dst := filepath.Join(dir, project.YamlRelPath)
	b, e := os.ReadFile(dst)
	if e != nil {
		log.Panic(e)
		project.LastError = e.Error()
		return
	}
	s := string(b)
	if !strings.Contains(s, tagReplacementIdentifier) {
		e = fmt.Errorf("invalid YAML file for k8s, tag identifier %s not exists at file: %s", tagReplacementIdentifier, project.YamlRelPath)
		log.Panic(e)
		project.LastError = e.Error()
		return
	}

	s = strings.Replace(s, tagReplacementIdentifier, tag, 1)
	e = os.WriteFile(dst, []byte(s), 0600)
	if e != nil {
		log.Panic(e)
		project.LastError = e.Error()
		return
	}

	cmd := exec.Command("kubectl", "apply", "-f", dst)
	cmd.Stderr = log.Writer()
	cmd.Stdout = log.Writer()
	e = cmd.Run()
	if e != nil {
		log.Panic(e)
		project.LastError = e.Error()
		return
	}
}

func parseTags(s string) (out string, e error) {
	var err = fmt.Errorf("invalid tags parameter:" + s)
	if s == "" {
		e = err
		return
	}

	ss := strings.Split(s, ",")
	for _, s := range ss {
		s = strings.Trim(s, " ")
		if strings.Contains(s, "sha-") {
			return s, nil
		}
		out = s
	}

	if s == "" {
		e = err
	}
	return
}
