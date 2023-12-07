package main

import (
	"embed"
	"html/template"
	"log"
)

var (
	//go:embed views
	fs embed.FS
	t  = template.Must(template.ParseFS(fs, "views/*.html"))
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime)
}
