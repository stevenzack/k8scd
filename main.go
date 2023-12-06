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
	port           = flag.String("p", ":9876", "The port you want to listen")
	logFile        = flag.String("l", "log.txt", "The log.txt file path")
	remoteIPHeader = flag.String("ip-header", "", "The header that contains remote IP, e.g 'X-Ip'. Default empty")
	kvStoreDir     = flag.String("kv", "", "Directory that store all the sensitive configuration data")
	adminPassword  string
)

var (
	stores *store.Store
)

func main() {
	flag.Parse()
	initialize()
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
	r.HandleFunc("/projects/", project)
	r.HandleFunc(dockerHubWebhookRoutePath, dockerhubWebhook)
	r.HandleFunc(apiNotifierRoutePath, notifier)
	log.Println("started http://localhost" + *port)
	println("started http://localhost" + *port)
	println("kv dir: ", *kvStoreDir)
	e = http.ListenAndServe(*port, r)
	if e != nil {
		log.Panic(e)
		return
	}
}

func initialize() {
	if kvStoreDir == nil || *kvStoreDir == "" {
		*kvStoreDir = os.Getenv("KV")
		if *kvStoreDir == "" {
			*kvStoreDir = "kv"
		}
	}

	b, e := os.ReadFile(filepath.Join(*kvStoreDir, pwdTxt))
	if e != nil {
		if os.IsNotExist(e) {
			return
		}
		log.Panic(e)
		return
	}
	s := string(b)
	if s != "" {
		adminPassword = s
	}
}
