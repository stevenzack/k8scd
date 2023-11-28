package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type (
	Repo struct {
		Id          string
		Name        string
		GitURL      string
		GitBranch   string
		YamlRelPath string // k8s/todo.yaml
		TagPrefix   string // sha-

		RunningTag string
		LastError  string

		UpdatedAt string
		CreatedAt string
	}
)

func ValidateRepoForm(r *http.Request) (v Repo, e error) {
	v.Name, e = parseFormValue(r, "name")
	if e != nil {
		return
	}
	v.GitURL, e = parseFormValue(r, "giturl")
	if e != nil {
		return
	}
	v.GitBranch, e = parseFormValue(r, "gitbranch")
	if e != nil {
		return
	}
	v.YamlRelPath, e = parseFormValue(r, "yamlrelpath")
	if e != nil {
		return
	}
	v.TagPrefix, e = parseFormValue(r, "tagprefix")
	if e != nil {
		return
	}

	return v, nil
}
func parseFormValue(r *http.Request, key string) (string, error) {
	v := r.FormValue(key)
	if v == "" {
		return "", fmt.Errorf("field " + key + " cannot be empty")
	}
	return v, nil
}

func (s *Repo) CloneGitRepoTo(dst string) error {
	os.MkdirAll(filepath.Dir(dst), 0700)

	cmd := exec.Command("git", "clone", "-b", s.GitBranch, s.GitURL, dst)
	cmd.Stderr = log.Writer()
	cmd.Stdout = log.Writer()
	return cmd.Run()
}
func (s *Store) GetRepos() ([]Repo, error) {
	b, e := s.getValue("repos")
	if e != nil {
		return nil, e
	}
	if len(b) == 0 {
		return nil, nil
	}
	var repos []Repo
	e = json.Unmarshal(b, &repos)
	if e != nil {
		return nil, e
	}

	return repos, nil
}

func (s *Store) GetRepoById(id string) (r Repo, e error) {
	vs, e := s.GetRepos()
	if e != nil {
		return
	}
	for _, v := range vs {
		if v.Id == id {
			r = v
			return
		}
	}
	e = fmt.Errorf("project not found:" + id)
	return
}

func (s *Store) SetRepos(repos []Repo) error {
	return s.setValue("repos", repos)
}

func (s *Store) UpdateRepo(v Repo) error {
	vs, e := s.GetRepos()
	if e != nil {
		return e
	}
	for i, v2 := range vs {
		if v2.Id == v.Id {
			v.UpdatedAt = time.Now().Format(time.DateTime)
			vs[i] = v
			return s.SetRepos(vs)
		}
	}
	return sql.ErrNoRows
}
func (s *Store) Insert(v Repo) error {
	if v.Id == "" {
		return fmt.Errorf("repo.Id cannot be empty")
	}

	vs, e := s.GetRepos()
	if e != nil {
		return e
	}
	for _, v2 := range vs {
		if v2.Id == v.Id {
			return fmt.Errorf("duplicated id for repo:" + v.Id)
		}
	}
	vs = append(vs, v)
	return s.SetRepos(vs)
}

func (s *Store) Delete(id string) error {
	var out []Repo
	vs, e := s.GetRepos()
	if e != nil {
		return e
	}
	for _, v := range vs {
		if v.Id != id {
			out = append(out, v)
		}
	}
	return s.SetRepos(out)
}
