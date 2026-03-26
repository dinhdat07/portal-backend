package types

import (
	"strings"
	"time"
)

const dateOnlyLayout = "2006-01-02"

type DateOnly struct {
	time.Time
}

func (d *DateOnly) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)

	t, err := time.Parse(dateOnlyLayout, s)
	if err != nil {
		return err
	}

	d.Time = t
	return nil
}

func (d DateOnly) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Format(dateOnlyLayout) + `"`), nil
}
