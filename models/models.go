package models

import "time"

type Prescription struct {
	ID            string
	Medication    Medication
	Schedule      Schedule
	Doses         int
	Refills       int
	ScheduleStart *time.Time
}

type Medication struct {
	ID      string
	Name    string
	Generic bool
	Brand   string
}

type Regimen struct {
	ID         string
	Medication Medication
	Doses      []Dose
	PatientID  string
}

type Dose struct {
	ID        string
	Time      time.Time
	Amount    float64
	Unit      string
	Taken     *bool
	Refill    int
	TimeTaken *time.Time
}

type ScheduledDose struct {
	DurationIntoPeriod Duration
	Amount             float64
	Unit               string
}

type Schedule struct {
	// Per period, the doses restart. No Dose Duration Into Period
	// can exceed the Period
	Period Duration
	Doses  []ScheduledDose
}

type User struct {
	ID   string
	Name string
}
