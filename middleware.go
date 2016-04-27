package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/zenazn/goji/web"
)

func BasicAuth(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Basic ") {
			log.Print("[INFO] Basic Header does not exist")
			http.Error(w, http.StatusText(404), 404)
			return
		}

		// NOTE: remove top of "Basic "
		authString, err := base64.StdEncoding.DecodeString(auth[6:])
		if err != nil {
			log.Printf("[ERROR] Decode fail err: %s", err.Error())
			http.Error(w, http.StatusText(404), 404)
			return
		}

		authInfo := strings.Split(string(authString), ":")
		fmt.Println(authInfo)
		if authInfo[0] != os.Getenv("BASIC_AUTH_USER") && authInfo[1] != os.Getenv("BASIC_AUTH_PASSWORD") {
			log.Print("[INFO] The user name or password unmatch")
			http.Error(w, http.StatusText(404), 404)
			return
		}

		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
