package manager

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/kzs0/kokoro/koko"
	"github.com/kzs0/kokoro/telemetry/metrics"
	"github.com/kzs0/pill_manager/models"
	"github.com/kzs0/pill_manager/models/repositories"
)

type Controller struct {
	Perscriptions *repositories.Perscriptions
	Users         *repositories.Users
	Handler       *Handler
}

// TODO add user information
func (c *Controller) GetRemainingDoses(w http.ResponseWriter, r *http.Request) {
	ctx, done := koko.Operation(r.Context(), "get_remaining_doses")
	var err error
	defer done(&ctx, &err)

	uid := r.PathValue("id")
	if uid == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(uid, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := c.Users.Get(ctx, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	doses, err := c.Handler.GetScheduledDoses(ctx, user)
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

	sid := r.PathValue("id")
	if sid == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(sid, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rx, err := c.Perscriptions.Get(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrPerscriptionNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
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

	id, err := strconv.ParseInt(uid, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := c.Users.Get(ctx, id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rx := &models.Perscription{}
	err = json.Unmarshal(payload, rx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO defaults
	if rx.Schedule.Period.Duration == 0 {
		rx.Schedule.Period = models.Duration{Duration: time.Hour * 24} // 1 day
	}

	err = c.Handler.NewPerscription(ctx, rx, user)
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

	_, err = c.Users.Put(ctx, user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	payload, err = json.Marshal(&user)
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

	sid := r.PathValue("id")
	if sid == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(sid, 10, 64)
	if err != nil {
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
