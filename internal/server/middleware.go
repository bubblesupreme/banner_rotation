package server

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		log.WithFields(log.Fields{
			"time:":  t.String(),
			"method": r.Method,
			"url":    r.URL.String(),
		}).Info("new request was handled")

		next.ServeHTTP(w, r)
	})
}

func jsonHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
