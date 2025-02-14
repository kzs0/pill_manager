package middleware

import (
	"log/slog"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/kzs0/pill_manager/models/db/sqlc"
)

func BlockUnapprovedUsers(next http.Handler, queries *sqlc.Queries) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`Unauthorized`))
			return
		}

		slog.Info("User Interaction", "uid", claims.RegisteredClaims.Subject)

		uid := claims.RegisteredClaims.Subject
		user, err := queries.GetUser(r.Context(), uid)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`Unauthorized`))
			return
		}

		if !user.Approved {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`Unauthorized`))
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}
