package main

import (
	"fmt"
	"net/http"
)

func badRequest(w http.ResponseWriter, e any) {
	http.Error(w, fmt.Sprint(e), http.StatusBadRequest)
}
