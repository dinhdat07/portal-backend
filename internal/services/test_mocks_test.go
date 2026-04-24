package services

import "time"

func cloneUser[T any](v *T) *T {
	if v == nil {
		return nil
	}
	cpy := *v
	return &cpy
}

func ptrString(s string) *string {
	return &s
}

func ptrTime(v time.Time) *time.Time {
	return &v
}
