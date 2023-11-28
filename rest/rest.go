package rest

import (
	"log"
	"net/http"
	"runtime/debug"
)

type RestServer struct {
	*http.ServeMux
}

func New() *RestServer {
	return &RestServer{
		ServeMux: http.DefaultServeMux,
	}
}

func (s *RestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			log.Println(e)
			debug.PrintStack()
			return
		}
	}()

	s.ServeMux.ServeHTTP(w, r)
}
