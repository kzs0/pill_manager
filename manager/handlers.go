package manager

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/kzs0/kokoro/koko"
	"github.com/kzs0/pill_manager/models"
	"github.com/kzs0/pill_manager/models/db/sqlc"
)

type Handler struct {
	Queries *sqlc.Queries
}

func (h *Handler) NewPerscription(ctx context.Context, rx *models.Prescription, user *models.User) (err error) {
	ctx, done := koko.Operation(ctx, "handler_new_rx")
	defer done(&ctx, &err)

	sch, err := json.Marshal(&rx.Schedule)
	if err != nil {
		return err
	}

	params := sqlc.CreateRxParams{
		ID:             uuid.NewString(),
		MedicationID:   rx.Medication.ID,
		ScheduledStart: sql.NullInt64{Int64: rx.ScheduleStart.Unix()},
		Refills:        int64(rx.Refills),
		Doses:          int64(rx.Doses),
		Schedule:       sch,
	}
	_, err = h.Queries.CreateRx(ctx, params)
	if err != nil {
		return err
	}

	regimenParams := sqlc.CreateRegimenParams{
		ID:           uuid.NewString(),
		MedicationID: rx.Medication.ID,
		Patient:      user.ID,
	}
	regimen, err := h.Queries.CreateRegimen(ctx, regimenParams)
	if err != nil {
		return err
	}

	t := rx.ScheduleStart
	// +1 because we want i to not cause a multiplication to 0
	for i := 1; i <= rx.Refills+1; i++ {
		for j := 0; j < rx.Doses; {
			for _, dose := range rx.Schedule.Doses {
				if j >= rx.Doses {
					break
				}

				dosesParams := sqlc.CreateDoseParams{
					ID:        uuid.NewString(),
					RegimenID: regimen.ID,
					Time:      t.Add(time.Duration(i) * dose.DurationIntoPeriod.Duration).Unix(),
					Amount:    dose.Amount,
					Unit:      dose.Unit,
				}
				_, err = h.Queries.CreateDose(ctx, dosesParams)
				if err != nil {
					return err
				}

				j++
			}

			t = t.Add(rx.Schedule.Period.Duration)
		}
	}

	return nil
}

func (h *Handler) MarkDoseTaken(ctx context.Context, id string) (err error) {
	ctx, done := koko.Operation(ctx, "handler_mark_dose_taken")
	defer done(&ctx, &err)

	params := sqlc.MarkDoseTakenParams{
		Taken:     sql.NullBool{Bool: true},
		TimeTaken: sql.NullInt64{Int64: time.Now().Unix()},
		ID:        id,
	}

	err = h.Queries.MarkDoseTaken(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) GetScheduledDoses(ctx context.Context, user *models.User) (_ []*models.Dose, err error) {
	ctx, done := koko.Operation(ctx, "handler_get_doses")
	defer done(&ctx, &err)

	rows, err := h.Queries.GetDosesByPatient(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	doses := make([]*models.Dose, 0, 3)
	for _, row := range rows {
		// Within a day of the dose scheduled time show
		// TODO make configurable
		if !row.Taken.Valid && time.Unix(row.Time, 0).Add(time.Hour*24).After(now) {
			dose := models.Dose{
				ID:     row.ID,
				Time:   time.Unix(row.Time, 0),
				Amount: row.Amount,
				Unit:   row.Unit,
			}
			doses = append(doses, &dose)
		}

		if len(doses) >= 3 {
			break
		}
	}

	return doses, nil
}
