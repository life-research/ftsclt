package models

import (
	"encoding/json"
	"time"
)

// ProcessStatus represents the status of a data processing job
type ProcessStatus struct {
	ProcessID           string    `json:"processId"`
	Phase               string    `json:"phase"`
	CreatedAt           time.Time `json:"-"` // Custom unmarshaling
	FinishedAt          time.Time `json:"-"` // Custom unmarshaling
	TotalPatients       int       `json:"totalPatients"`
	TotalBundles        int       `json:"totalBundles"`
	DeidentifiedBundles int       `json:"deidentifiedBundles"`
	SentBundles         int       `json:"sentBundles"`
	SkippedBundles      int       `json:"skippedBundles"`
}

// Custom unmarshaling for the timestamp arrays
func (p *ProcessStatus) UnmarshalJSON(data []byte) error {
	type Alias ProcessStatus
	aux := &struct {
		CreatedAt  []int `json:"createdAt"`
		FinishedAt []int `json:"finishedAt"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert array to time.Time for CreatedAt
	if len(aux.CreatedAt) == 7 {
		p.CreatedAt = time.Date(
			aux.CreatedAt[0],             // year
			time.Month(aux.CreatedAt[1]), // month
			aux.CreatedAt[2],             // day
			aux.CreatedAt[3],             // hour
			aux.CreatedAt[4],             // minute
			aux.CreatedAt[5],             // second
			aux.CreatedAt[6],             // nanosecond
			time.UTC,
		)
	}

	// Convert array to time.Time for FinishedAt
	if len(aux.FinishedAt) == 7 {
		p.FinishedAt = time.Date(
			aux.FinishedAt[0],
			time.Month(aux.FinishedAt[1]),
			aux.FinishedAt[2],
			aux.FinishedAt[3],
			aux.FinishedAt[4],
			aux.FinishedAt[5],
			aux.FinishedAt[6],
			time.UTC,
		)
	}

	return nil
}
