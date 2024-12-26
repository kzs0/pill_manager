package models

import "time"

type PerscriptionOpts struct {
	SDO *ScheduledDoseOpts
}

type ScheduledDoseOpts struct {
	Taken *bool
	Time  *time.Time
}
