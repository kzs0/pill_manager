package middleware

import (
	"net/http"
	"strings"
)

type CORSOptions struct {
	Origin  []string
	Methods []string
	Headers []string
}

func CORS(next http.Handler, opts *CORSOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origins := strings.Join(opts.Origin, ",")
		methods := strings.Join(opts.Methods, ",")
		headers := strings.Join(opts.Headers, ",")

		w.Header().Set("Access-Control-Allow-Origin", origins)
		w.Header().Set("Access-Control-Allow-Methods", methods)
		w.Header().Set("Access-Control-Allow-Headers", headers)

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	}
}
