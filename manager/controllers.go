package manager

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

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

// TODO add user information
func (c *Controller) GetRemainingDoses(w http.ResponseWriter, r *http.Request) {
	ctx, done := koko.Operation(r.Context(), "get_remaining_doses")
	var err error
	defer done(&ctx, &err)

	id := r.PathValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	doses, err := c.Handler.GetScheduledDoses(ctx, id)
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

	uid := r.Header.Get("uid")
	if uid == "" {
		w.WriteHeader(http.StatusBadRequest)
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

	params := sqlc.CreateUserParams{
		ID:   uuid.NewString(),
		Name: user.Name,
	}

	userdb, err := c.Queries.CreateUser(ctx, params)
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
	ctx, done := koko.Operation(r.Context(), "post_perscription")
	var err error
	defer done(&ctx, &err)

	id := r.PathValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = c.Handler.MarkDoseTaken(ctx, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (c *Controller) GetRoot(w http.ResponseWriter, r *http.Request) {
	ctx, done := koko.Operation(r.Context(), "get_root", metrics.WithLabelNames("test"))
	var err error
	defer done(&ctx, &err)

	ctx = koko.Register(ctx, koko.Str("test", "test1"))

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
