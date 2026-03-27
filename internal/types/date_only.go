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
	return d.parse(s)
}

func (d DateOnly) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Format(dateOnlyLayout) + `"`), nil
}

func (d *DateOnly) UnmarshalText(text []byte) error {
	return d.parse(string(text))
}

func (d *DateOnly) parse(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		d.Time = time.Time{}
		return nil
	}

	t, err := time.Parse(dateOnlyLayout, s)
	if err != nil {
		return err
	}

	d.Time = t
	return nil
}
