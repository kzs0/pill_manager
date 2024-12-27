-- name: CreateRegimen :one
INSERT INTO regimens (id, medication_id, patient, prescription_id)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: CreateDose :one
INSERT INTO doses (id, regimen_id, time, amount, unit)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetDosesByPatient :many
SELECT doses.*, medications.*, regimens.* FROM doses
INNER JOIN regimens ON doses.regimen_id = regimens.id
INNER JOIN medications ON regimens.medication_id = medications.id
WHERE regimens.patient = ?
ORDER BY doses.Time;

-- name: MarkDoseTaken :exec
UPDATE doses
SET taken = ?, time_taken = ?
WHERE id = ?;

-- name: CreateRx :one
INSERT INTO prescriptions (id, medication_id, scheduled_start, refills, doses, schedule, patient)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetRx :one
SELECT * FROM prescriptions
WHERE id = ?;

-- name: CreateMedication :one
INSERT INTO medications (id, name, generic, brand)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetMedication :one
SELECT * FROM medications
WHERE id = ?;

-- name: CreateUser :one
INSERT INTO users (id)
VALUES (?)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = ?;
