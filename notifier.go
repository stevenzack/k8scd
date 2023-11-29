package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/stevenzack/k8scd/store"
)

const (
	apiNotifierRoutePath     = "/api/notifier/"
	tagReplacementIdentifier = "{{.Tag}}"
)

// tags: name/app:latest\nname/app:1.0.0
func notifier(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) <= len(apiNotifierRoutePath) {
		return
	}
	id := r.URL.Path[len(apiNotifierRoutePath):]
	if id == "" || strings.Contains(id, "|") {
		badRequest(w, "project id empty or invalid")
		return
	}

	e := r.ParseMultipartForm(24 << 20)
	if e != nil {
		badRequest(w, e.Error())
		return
	}

	tag, e := parseTags(r.MultipartForm.Value["tags"])
	if e != nil {
		badRequest(w, e.Error())
		return
	}
	data := map[string]any{
		"Tag": tag,
	}
	for k, vs := range r.MultipartForm.Value {
		var value = ""
		if len(vs) > 0 {
			value = vs[0]
		}

		data[k] = value
	}

	project, e := stores.GetRepoById(id)
	if e != nil {
		badRequest(w, "project id empty or invalid:"+id)
		return
	}
	defer stores.UpdateRepo(project)

	//do
	dir := filepath.Join("notifier-cache", project.Id)
	// defer os.RemoveAll(dir)
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

	t, e := template.New(project.YamlRelPath).Parse(s)
	if e != nil {
		log.Panic(e)
		project.LastError = e.Error()
		return
	}
	fo, e := os.OpenFile(dst, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0600)
	if e != nil {
		log.Panic(e)
		project.LastError = e.Error()
		return
	}
	e = t.Execute(fo, data)
	if e != nil {
		fo.Close()
		log.Panic(e)
		project.LastError = e.Error()
		return
	}
	fo.Close()

	cmd := exec.Command("kubectl", "apply", "-f", dst)
	cmd.Stderr = log.Writer()
	cmd.Stdout = log.Writer()
	e = cmd.Run()
	if e != nil {
		log.Panic(e)
		project.LastError = e.Error()
		return
	}
	project.RunningTag = tag
	e = stores.InsertDeployment(store.Deployment{
		Id:        strconv.FormatInt(time.Now().Unix(), 10),
		ProjectId: id,
		Tag:       tag,
		CreatedAt: time.Now().Format(time.DateTime),
	})
	if e != nil {
		log.Println(e)
		return
	}
}

// zigzigcheers/todo:main \n, zigzigcheers/todo:sha-82e4bb3
func parseTags(tags []string) (out string, e error) {
	var s = ""
	if len(tags) > 0 {
		s = tags[0]
	}
	var err = fmt.Errorf("invalid tags parameter:" + s)
	if s == "" {
		e = err
		return
	}

	s = strings.ReplaceAll(s, "\n", ",")
	ss := strings.Split(s, ",")
	for _, s := range ss {
		s = strings.Trim(s, " ")
		s = SubAfter(s, ":", s)
		if strings.HasPrefix(s, "sha-") {
			return s, nil
		}
		out = s
	}

	if s == "" {
		e = err
	}
	return
}

func SubAfter(s, sep, def string) string {
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			return s[i+len(sep):]
		}
	}
	return def
}
