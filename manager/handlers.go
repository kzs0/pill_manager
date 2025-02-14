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

func (h *Handler) NewPerscription(ctx context.Context, rx *models.Prescription, uid string) (_ *models.Prescription, err error) {
	ctx, done := koko.Operation(ctx, "handler_new_rx")
	defer done(&ctx, &err)

	medicationParams := sqlc.CreateMedicationParams{
		ID:      uuid.NewString(),
		Name:    rx.Medication.Name,
		Generic: rx.Medication.Generic,
		Brand:   rx.Medication.Brand,
	}
	medication, err := h.Queries.CreateMedication(ctx, medicationParams)
	if err != nil {
		return nil, err
	}

	sch, err := json.Marshal(&rx.Schedule)
	if err != nil {
		return nil, err
	}

	params := sqlc.CreateRxParams{
		ID:             uuid.NewString(),
		MedicationID:   medication.ID,
		ScheduledStart: sql.NullInt64{Int64: rx.ScheduleStart.Unix()},
		Refills:        int64(rx.Refills),
		Doses:          int64(rx.Doses),
		Schedule:       sch,
		Patient:        uid,
	}
	prescription, err := h.Queries.CreateRx(ctx, params)
	if err != nil {
		return nil, err
	}

	regimenParams := sqlc.CreateRegimenParams{
		ID:             uuid.NewString(),
		MedicationID:   medication.ID,
		Patient:        uid,
		PrescriptionID: prescription.ID,
	}
	regimen, err := h.Queries.CreateRegimen(ctx, regimenParams)
	if err != nil {
		return nil, err
	}

	t := rx.ScheduleStart
	for i := 0; i <= rx.Refills; i++ {
		for j := 0; j < rx.Doses; {
			for _, dose := range rx.Schedule.Doses {
				if j >= rx.Doses {
					break
				}

				doseTime := t.Add(dose.DurationIntoPeriod.Duration)
				dosesParams := sqlc.CreateDoseParams{
					ID:        uuid.NewString(),
					RegimenID: regimen.ID,
					Refill:    int64(i),
					Time:      doseTime.Unix(),
					Amount:    dose.Amount,
					Unit:      dose.Unit,
				}
				_, err = h.Queries.CreateDose(ctx, dosesParams)
				if err != nil {
					return nil, err
				}

				j++
			}

			ttemp := t.Add(rx.Schedule.Period.Duration)
			t = &ttemp
		}
	}

	var schedule models.Schedule
	err = json.Unmarshal(prescription.Schedule, &schedule)
	if err != nil {
		return nil, err
	}

	var start *time.Time
	if prescription.ScheduledStart.Valid {
		starttemp := time.Unix(prescription.ScheduledStart.Int64, 0)
		start = &starttemp
	}

	rx = &models.Prescription{
		ID: prescription.ID,
		Medication: models.Medication{
			ID:      medication.ID,
			Name:    medication.Name,
			Generic: medication.Generic,
			Brand:   medication.Brand,
		},
		Doses:         int(prescription.Doses),
		Refills:       int(prescription.Refills),
		Schedule:      schedule,
		ScheduleStart: start,
	}

	return rx, nil
}

func (h *Handler) MarkDoseTaken(ctx context.Context, id string, taken bool, t time.Time) (err error) {
	ctx, done := koko.Operation(ctx, "handler_mark_dose_taken")
	defer done(&ctx, &err)

	params := sqlc.MarkDoseTakenParams{
		Taken:     sql.NullBool{Bool: taken, Valid: true},
		TimeTaken: sql.NullInt64{Int64: t.Unix(), Valid: true},
		ID:        id,
	}

	err = h.Queries.MarkDoseTaken(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) GetScheduledDoses(ctx context.Context, uid string, limit int) (_ []models.Regimen, err error) {
	ctx, done := koko.Operation(ctx, "handler_get_doses")
	defer done(&ctx, &err)

	args := sqlc.GetDosesByPatientLimitByParams{
		Patient: uid,
		Limit:   int64(limit),
	}

	rows, err := h.Queries.GetDosesByPatientLimitBy(ctx, args)
	if err != nil {
		return nil, err
	}

	regimenMap := make(map[string]*models.Regimen, 0)
	for _, row := range rows {
		regimen, ok := regimenMap[row.ID_3]
		if !ok {
			med := models.Medication{
				ID:      row.ID_2,
				Name:    row.Name,
				Generic: row.Generic,
				Brand:   row.Brand,
			}

			regimen = &models.Regimen{
				ID:         row.ID_3,
				PatientID:  uid,
				Medication: med,
				Doses:      make([]models.Dose, 0, 0),
			}

			regimenMap[row.ID_3] = regimen
		}

		if row.Taken.Valid {
			continue // skip
		}

		dose := models.Dose{
			ID:     row.ID,
			Time:   time.Unix(row.Time, 0),
			Amount: row.Amount,
			Unit:   row.Unit,
			Refill: int(row.Refill),
		}

		regimen.Doses = append(regimen.Doses, dose)
	}

	regimens := make([]models.Regimen, 0, 0)
	for _, regimen := range regimenMap {
		regimens = append(regimens, *regimen)
	}

	return regimens, nil
}
