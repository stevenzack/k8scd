package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func badRequest(w http.ResponseWriter, e any) {
	http.Error(w, fmt.Sprint(e), http.StatusBadRequest)
}

func newID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}
func validateID(s string) bool {
	return len(s) == 32
}
