package middleware

import (
	"context"
	"net/http"

	"github.com/kzs0/kokoro/koko"
	"github.com/kzs0/kokoro/telemetry/metrics"
)

func HttpOperation(ctx context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, done := koko.Operation(ctx, "http", metrics.WithLabelNames("abc"))
		defer done(&ctx, nil)

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
