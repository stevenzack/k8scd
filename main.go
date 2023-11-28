package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/stevenzack/k8scd/rest"
	"github.com/stevenzack/k8scd/store"
)

var (
	port           = flag.String("p", ":8080", "The port you want to listen")
	logFile        = flag.String("l", "log.txt", "The log.txt file path")
	adminPassword  = flag.String("pwd", "", "admin password")
	remoteIPHeader = flag.String("ip-header", "", "The header that contains remote IP, e.g 'X-Ip'. Default empty")
	kvStoreDir     = flag.String("dir", "db", "Directory that store all the sensitive configuration data")
)

var (
	stores *store.Store
)

func main() {
	flag.Parse()
	if adminPassword == nil || *adminPassword == "" || len(*adminPassword) < 6 {
		log.Panic("admin password not or invalid")
		return
	}
	e := os.MkdirAll(filepath.Dir(*logFile), 0755)
	if e != nil {
		log.Panic(e)
		return
	}

	fo, e := os.OpenFile(*logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if e != nil {
		log.Panic(e)
		return
	}
	defer fo.Close()
	log.SetOutput(fo)

	stores, e = store.NewStore(*kvStoreDir)
	if e != nil {
		log.Panic(e)
		return
	}

	//run
	r := rest.New()
	r.HandleFunc("/", home)
	r.HandleFunc("/login", login)
	r.HandleFunc(dockerHubWebhookRoutePath, dockerhubWebhook)
	// r.HandleFunc(apiNotifierRoutePath, notifier)
	log.Println("started http://localhost" + *port)
	e = http.ListenAndServe(*port, r)
	if e != nil {
		log.Panic(e)
		return
	}
}
