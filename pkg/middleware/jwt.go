package middleware

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

type Auth0Config struct {
	Audience string `env:"AUTH0_AUDIENCE" envDefault:"https://kenzo.us.auth0.com/api/v2/"`
	Domain   string `env:"AUTH0_DOMAIN" envDefault:"kenzo.us.auth0.com"`
}

type CustomClaims struct{}

func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

func EnsureValidToken(next http.Handler, cfg Auth0Config) http.Handler {
	issuerURL, err := url.Parse("https://" + cfg.Domain + "/")
	if err != nil {
		log.Fatalf("Failed to parse the issuer url: %v", err)
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{cfg.Audience},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &CustomClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to set up the jwt validator")
	}

	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Encountered error while validating JWT: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Failed to validate JWT."}`))
	}

	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
	)

	return middleware.CheckJWT(next)
}

// // HasScope checks whether our claims have a specific scope.
// func (c CustomClaims) HasScope(expectedScope string) bool {
// 	result := strings.Split(c.Scope, " ")
// 	for i := range result {
// 		if result[i] == expectedScope {
// 			return true
// 		}
// 	}

// 	return false
// }
