package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	loginCache    = make(map[string]time.Time)
	jwtSecret     = []byte("hGLXkPLD48AD")
	jwtCookieName = "at"
	jwtAge        = time.Hour * 24
	pwdTxt        = "pwd.txt"
)

func login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if adminPassword == "" {
			adminPassword = newID()
			e := os.WriteFile(filepath.Join(*kvStoreDir, pwdTxt), []byte(adminPassword), 0600)
			if e != nil {
				log.Panic(e)
				return
			}

			t.ExecuteTemplate(w, "login.html", adminPassword)
			return
		}
		t.ExecuteTemplate(w, "login.html", "")
		return
	case http.MethodPost:
		ip := r.RemoteAddr
		if remoteIPHeader != nil && *remoteIPHeader != "" {
			ip = r.Header.Get(*remoteIPHeader)
		}
		last, ok := loginCache[ip]
		if ok && last.Add(time.Minute*5).After(time.Now()) {
			http.Error(w, "too frequent request", http.StatusBadRequest)
			return
		}
		loginCache[ip] = time.Now()

		if r.FormValue("password") != adminPassword {
			time.Sleep(time.Second * 5)
			http.Error(w, "password incorrect", http.StatusBadRequest)
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"exp": time.Now().Add(jwtAge).Unix(),
		})
		// Sign and get the complete encoded token as a string using the secret
		tokenString, e := token.SignedString(jwtSecret)
		if e != nil {
			log.Panic(e)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     jwtCookieName,
			Value:    tokenString,
			Expires:  time.Now().Add(jwtAge),
			MaxAge:   int(jwtAge.Seconds()),
			HttpOnly: true,
		})
		fmt.Fprint(w, "/")
		return

	}
}

func auth(w http.ResponseWriter, r *http.Request) bool {
	c, e := r.Cookie(jwtCookieName)
	if e != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return true
	}
	_, e = jwt.Parse(c.Value, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return jwtSecret, nil
	})
	if e != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return true
	}
	return false
}
