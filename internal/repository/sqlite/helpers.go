package sqlite

import "time"

const timeFormat = time.RFC3339

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
