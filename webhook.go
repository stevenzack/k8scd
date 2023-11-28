package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/stevenzack/k8scd/store"
)

type WebHookRequest struct {
	CallbackUrl string `json:"callback_url"`
	PushData    struct {
		Pusher   string `json:"pusher"`    //zigzigcheers
		PushedAt int64  `json:"pushed_at"` //1701171615
		Tag      string `json:"tag"`       // main sha-13hjg4
	} `json:"push_data"`
	Repository struct {
		Status          string  `json:"status"`    // Active
		Namespace       string  `json:"namespace"` //zigzigcheers
		Name            string  `json:"name"`      //todo
		RepoName        string  `json:"repo_name"` // zigzigcheers/todo
		RepoUrl         string  `json:"repo_url"`  // https://hub.docker.com/r/zigzigcheers/todo
		Description     string  `json:"description"`
		FullDescription *string `json:"full_description"`
		StarCount       int     `json:"star_count"`
		IsPrivate       bool    `json:"is_private"`
		IsTrusted       bool    `json:"is_trusted"`
		IsOfficial      bool    `json:"is_official"`
		Owner           string  `json:"owner"`
		DateCreated     int64   `json:"date_created"`
	} `json:"repository"`
}
type CallbackState string

const (
	CallbackStateSuccess CallbackState = "success"
	CallbackStateFailure CallbackState = "failure"
	CallbackStateError   CallbackState = "error"
)

type WebHookCallback struct {
	State       CallbackState `json:"state"`       // Accepted values are success, failure, and error. If the state isn't success, the webhook chain is interrupted.
	Description string        `json:"description"` //
	Context     string        `json:"context"`     // "Continuous integration by Acme CI", A string containing the context of the operation. Can be retrieved from the Docker Hub. Maximum 100 characters
	TargetUrl   string        `json:"target_url"`  // "https://ci.acme.com/results/afd339c1c3d27", The URL where the results of the operation can be found. Can be retrieved on the Docker Hub.
}

const (
	dockerHubWebhookRoutePath = "/api/docker-hub-webhook/"
)

func dockerhubWebhook(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) <= len(dockerHubWebhookRoutePath) {
		return
	}
	id := r.URL.Path[len(dockerHubWebhookRoutePath):]
	if id == "" || strings.Contains(id, "|") {
		badRequest(w, "project id empty or invalid")
		return
	}
	project, e := stores.GetRepoById(id)
	if e != nil {
		badRequest(w, e.Error())
		return
	}

	defer r.Body.Close()
	b, e := io.ReadAll(r.Body)
	if e != nil {
		log.Panic(e)
		return
	}
	var req WebHookRequest
	e = json.Unmarshal(b, &req)
	if e != nil {
		log.Panic(e)
		return
	}
	// get tag
	if !strings.HasPrefix(req.PushData.Tag, project.TagPrefix) {
		log.Println("ignored tag change:", req.PushData.Tag, " for project "+project.Name)
		return
	}

	//do the job
	defer stores.UpdateRepo(project)
	dir := filepath.Join("docker-webhook-cache", project.Id)
	defer os.RemoveAll(dir)
	e = project.CloneGitRepoTo(dir)
	if e != nil {
		log.Panic(e)
		project.LastError = e.Error()
		return
	}

	rel := project.YamlRelPath
	dst := filepath.Join(dir, rel)
	b, e = os.ReadFile(dst)
	if e != nil {
		log.Panic(e)
		project.LastError = e.Error()
		return
	}
	s := string(b)
	if !strings.Contains(s, tagReplacementIdentifier) {
		e = fmt.Errorf("invalid YAML file for k8s, tag identifier %s not exists at file: %s", tagReplacementIdentifier, rel)
		log.Panic(e)
		project.LastError = e.Error()
		return
	}

	s = strings.Replace(s, tagReplacementIdentifier, req.PushData.Tag, 1)
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
	b, e = json.Marshal(WebHookCallback{
		State: CallbackStateSuccess,
	})
	if e != nil {
		log.Panic(e)
		return
	}

	e = stores.InsertDeployment(store.Deployment{
		Id:        strconv.FormatInt(time.Now().Unix(), 10),
		ProjectId: id,
		Tag:       req.PushData.Tag,
		CreatedAt: time.Now().Format(time.DateTime),
	})
	if e != nil {
		log.Println(e)
		return
	}

	if req.CallbackUrl == "" {
		return
	}
	go func() {
		_, e = http.Post(req.CallbackUrl, "application/json", bytes.NewReader(b))
		if e != nil {
			log.Println(e)
			return
		}
	}()
}
