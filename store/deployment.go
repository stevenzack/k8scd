package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

type (
	Deployment struct {
		Id        string
		ProjectId string
		Tag       string
		CreatedAt string
	}
)

func (d *Deployment) getStoreKey() string {
	return "repos/" + d.ProjectId + "/" + d.Id + ".json"
}

func (s *Store) InsertDeployment(v Deployment) error {
	return s.setValue(v.getStoreKey(), v)
}

func (s *Store) QueryDeployment(projectId string) ([]Deployment, error) {
	dir := filepath.Join(s.dir, "repos", projectId)
	list, e := os.ReadDir(dir)
	if e != nil {
		return nil, nil
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name() > list[j].Name()
	})

	var out []Deployment
	for _, f := range list {
		b, e := os.ReadFile(filepath.Join(dir, f.Name()))
		if e != nil {
			return nil, e
		}
		var v Deployment
		e = json.Unmarshal(b, &v)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, nil
}
