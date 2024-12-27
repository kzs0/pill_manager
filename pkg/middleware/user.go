package middleware

import (
	"database/sql"
	"errors"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/kzs0/pill_manager/models/db/sqlc"
)

func ObserveNewUsers(next http.Handler, queries *sqlc.Queries) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		uid := claims.RegisteredClaims.Subject

		_, err := queries.GetUser(r.Context(), uid)
		if errors.Is(err, sql.ErrNoRows) {
			_, err = queries.CreateUser(r.Context(), uid)
			if err != nil {
				http.Error(w, "failed to manage new user", http.StatusInternalServerError)
				return
			}
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}
