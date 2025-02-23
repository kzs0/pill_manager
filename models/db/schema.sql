CREATE TABLE doses (
    id TEXT PRIMARY KEY,
    regimen_id TEXT NOT NULL, -- References Regimen ID
    refill INT NOT NULL, -- which refill this dose is in
    time BIGINT NOT NULL, -- seconds since epoch
    amount REAL NOT NULL,
    unit TEXT NOT NULL,
    taken BOOLEAN, -- If Null, to be taken (hasn't confirmed that it was or wasn't)
    time_taken BIGINT, -- If Null, not determined taken or not
    FOREIGN KEY (regimen_id) REFERENCES regimens (id)
);

CREATE TABLE regimens (
    id TEXT PRIMARY KEY,
    medication_id TEXT NOT NULL, -- References Medication ID
    patient TEXT NOT NULL, -- References User ID
    prescription_id TEXT NOT NULL, -- References Prescription ID
    FOREIGN KEY (medication_id) REFERENCES medications (id),
    FOREIGN KEY (patient) REFERENCES users (id)
);

CREATE TABLE prescriptions (
    id TEXT PRIMARY KEY,
    medication_id TEXT NOT NULL, -- References Medication ID
    schedule BLOB NOT NULL, -- JSON schedule
    scheduled_start BIGINT, -- If Null, hasn't started
    refills INT NOT NULL,
    doses INT NOT NULL,
    patient TEXT NOT NULL, -- References User ID
    FOREIGN KEY (medication_id) REFERENCES medications (id),
    FOREIGN KEY (patient) REFERENCES users (id)
);

CREATE TABLE medications (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    generic BOOLEAN NOT NULL,
    brand TEXT NOT NULL
);

CREATE TABLE users (id TEXT PRIMARY KEY, approved BOOLEAN NOT NULL);
