package manager

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/google/uuid"
	"github.com/kzs0/kokoro/koko"
	"github.com/kzs0/kokoro/telemetry/metrics"
	"github.com/kzs0/pill_manager/models"
	"github.com/kzs0/pill_manager/models/db/sqlc"
)

type Controller struct {
	Queries *sqlc.Queries
	Handler *Handler
}

func (c *Controller) GetRemainingDoses(w http.ResponseWriter, r *http.Request) {
	ctx, done := koko.Operation(r.Context(), "get_remaining_doses")
	var err error
	defer done(&ctx, &err)

	claims, ok := ctx.Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
	if !ok {
		slog.Error("missing jwt claims in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	uid := claims.RegisteredClaims.Subject

	doses, err := c.Handler.GetScheduledDoses(ctx, uid, 1000)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	payload, err := json.Marshal(&doses)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (c *Controller) GetLimitedRemainingDoses(w http.ResponseWriter, r *http.Request) {
	ctx, done := koko.Operation(r.Context(), "get_limited_remaining_doses")
	var err error
	defer done(&ctx, &err)

	claims, ok := ctx.Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
	if !ok {
		slog.Error("missing jwt claims in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	uid := claims.RegisteredClaims.Subject

	countS := r.PathValue("count")
	if countS == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	count, err := strconv.ParseInt(countS, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	doses, err := c.Handler.GetScheduledDoses(ctx, uid, int(count))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	payload, err := json.Marshal(&doses)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (c *Controller) GetPerscription(w http.ResponseWriter, r *http.Request) {
	ctx, done := koko.Operation(r.Context(), "get_perscription")
	var err error
	defer done(&ctx, &err)

	id := r.PathValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rxdb, err := c.Queries.GetRx(ctx, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sch := models.Schedule{}
	err = json.Unmarshal(rxdb.Schedule, &sch)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var start *time.Time
	if rxdb.ScheduledStart.Valid {
		s := time.Unix(rxdb.ScheduledStart.Int64, 0)
		start = &s
	}

	medicationdb, err := c.Queries.GetMedication(ctx, rxdb.MedicationID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rx := models.Prescription{
		ID: rxdb.ID,
		Medication: models.Medication{
			ID:      medicationdb.ID,
			Name:    medicationdb.Name,
			Generic: medicationdb.Generic,
			Brand:   medicationdb.Brand,
		},
		Schedule:      sch,
		Doses:         int(rxdb.Doses),
		Refills:       int(rxdb.Refills),
		ScheduleStart: start,
	}

	payload, err := json.Marshal(&rx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(payload)
}

func (c *Controller) PostPerscription(w http.ResponseWriter, r *http.Request) {
	ctx, done := koko.Operation(r.Context(), "post_perscription")
	var err error
	defer done(&ctx, &err)

	claims, ok := ctx.Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
	if !ok {
		slog.Error("missing jwt claims in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	uid := claims.RegisteredClaims.Subject
	if uid == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rx := &models.Prescription{}
	err = json.Unmarshal(payload, rx)
	if err != nil {
		slog.Warn("failed to unmarshal rx", "err", err, "payload", string(payload))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO defaults
	if rx.Schedule.Period.Duration == 0 {
		rx.Schedule.Period = models.Duration{Duration: time.Hour * 24} // 1 day
	}

	rx, err = c.Handler.NewPerscription(ctx, rx, uid)
	if err != nil {
		slog.Error("err %+v", err)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	payload, err = json.Marshal(rx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(payload)
}

func (c *Controller) DosesTillEmpty(w http.ResponseWriter, r *http.Request) {
	ctx, done := koko.Operation(r.Context(), "doses_till_empty")
	var err error
	defer done(&ctx, &err)

	id := r.PathValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	doses, err := c.Queries.DosesTillEmpty(ctx, id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Write([]byte(fmt.Sprintf(`{"doses": %d}`, doses)))
}

func (c *Controller) DosesTillRefill(w http.ResponseWriter, r *http.Request) {
	ctx, done := koko.Operation(r.Context(), "doses_till_refill")
	var err error
	defer done(&ctx, &err)

	id := r.PathValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	arg := sqlc.DosesTillRefillParams{
		RegimenID:   id,
		RegimenID_2: id,
	}
	doses, err := c.Queries.DosesTillRefill(ctx, arg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Write([]byte(fmt.Sprintf(`{"doses": %d}`, doses)))
}

func (c *Controller) PostUser(w http.ResponseWriter, r *http.Request) {
	ctx, done := koko.Operation(r.Context(), "post_user")
	var err error
	defer done(&ctx, &err)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user := &models.User{}
	err = json.Unmarshal(payload, user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userdb, err := c.Queries.CreateUser(ctx, uuid.NewString())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	payload, err = json.Marshal(&userdb)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(payload)
}

func (c *Controller) PostTaken(w http.ResponseWriter, r *http.Request) {
	ctx, done := koko.Operation(r.Context(), "post_taken")
	var err error
	defer done(&ctx, &err)

	id := r.PathValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload := make(map[string]string, 1)
	err = json.Unmarshal(body, &payload)
	if err != nil {
		slog.Error("failed to unmarshal payload", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(payload) == 0 || len(payload) > 1 {
		slog.Warn("incorrect keys", "num_keys", len(payload))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	t, err := time.Parse(time.RFC3339, payload["time"])
	if err != nil {
		slog.Warn("failed to parse time", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = c.Handler.MarkDoseTaken(ctx, id, true, t)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (c *Controller) PostSkipped(w http.ResponseWriter, r *http.Request) {
	ctx, done := koko.Operation(r.Context(), "post_skipped")
	var err error
	defer done(&ctx, &err)

	id := r.PathValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload := make(map[string]string, 1)
	err = json.Unmarshal(body, &payload)
	if err != nil {
		slog.Error("failed to unmarshal payload", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(payload) == 0 || len(payload) > 1 {
		slog.Warn("incorrect keys", "num_keys", len(payload))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	t, err := time.Parse(time.RFC3339, payload["time"])
	if err != nil {
		slog.Warn("failed to parse time", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = c.Handler.MarkDoseTaken(ctx, id, false, t)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (c *Controller) GetRoot(w http.ResponseWriter, r *http.Request) {
	ctx, done := koko.Operation(r.Context(), "get_root", metrics.WithLabelNames("test"))
	var err error
	defer done(&ctx, &err)

	w.WriteHeader(http.StatusOK)
}

func (c *Controller) Options(w http.ResponseWriter, r *http.Request) {
	ctx, done := koko.Operation(r.Context(), "cors_options")
	var err error
	defer done(&ctx, &err)

	enableCORS(w, r)

	w.WriteHeader(http.StatusOK)
}

func enableCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")                                // Allow all origins
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS") // Allowed HTTP methods
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")     // Allowed headers
	w.Header().Set("Access-Control-Expose-Headers", "Content-Length")                 // Expose headers
	w.Header().Set("Access-Control-Allow-Credentials", "true")                        // Allow credentials
}
