package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
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

		UpdatedAt time.Time
		CreatedAt time.Time
	}
)

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
			v.UpdatedAt = time.Now()
			vs[i] = v
			return s.SetRepos(vs)
		}
	}
	return sql.ErrNoRows
}
