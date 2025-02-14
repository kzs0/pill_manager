-- name: CreateRegimen :one
INSERT INTO
    regimens (id, medication_id, patient, prescription_id)
VALUES
    (?, ?, ?, ?) RETURNING *;

-- name: CreateDose :one
INSERT INTO
    doses (id, regimen_id, refill, time, amount, unit)
VALUES
    (?, ?, ?, ?, ?, ?) RETURNING *;

-- name: DosesTillEmpty :one
SELECT
    count(*)
FROM
    doses
WHERE
    doses.regimen_id = ?
    AND doses.taken IS NULL;

-- name: DosesTillRefill :one
SELECT
    COUNT(*)
FROM
    doses
WHERE
    doses.regimen_id = ?
    AND doses.taken IS NULL
    AND doses.refill = (
        SELECT
            MIN(dose_inner.refill)
        FROM
            doses AS dose_inner
        WHERE
            dose_inner.regimen_id = ?
            AND dose_inner.taken IS NULL
    );

-- name: GetDosesByPatient :many
SELECT
    doses.*,
    medications.*,
    regimens.*
FROM
    doses
    INNER JOIN regimens ON doses.regimen_id = regimens.id
    INNER JOIN medications ON regimens.medication_id = medications.id
WHERE
    doses.taken IS NULL
    AND regimens.patient = ?
ORDER BY
    doses.Time;

-- name: GetDosesByPatientLimitBy :many
SELECT
    doses.*,
    medications.*,
    regimens.*
FROM
    doses
    INNER JOIN regimens ON doses.regimen_id = regimens.id
    INNER JOIN medications ON regimens.medication_id = medications.id
WHERE
    doses.taken IS NULL
    AND regimens.patient = ?
ORDER BY
    doses.Time
LIMIT
    ?;

-- name: MarkDoseTaken :exec
UPDATE doses
SET
    taken = ?,
    time_taken = ?
WHERE
    id = ?;

-- name: CreateRx :one
INSERT INTO
    prescriptions (
        id,
        medication_id,
        scheduled_start,
        refills,
        doses,
        schedule,
        patient
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: GetRx :one
SELECT
    *
FROM
    prescriptions
WHERE
    id = ?;

-- name: CreateMedication :one
INSERT INTO
    medications (id, name, generic, brand)
VALUES
    (?, ?, ?, ?) RETURNING *;

-- name: GetMedication :one
SELECT
    *
FROM
    medications
WHERE
    id = ?;

-- name: CreateUser :one
INSERT INTO
    users (id, approved)
VALUES
    (?, false) RETURNING *;

-- name: GetUser :one
SELECT
    *
FROM
    users
WHERE
    id = ?;
