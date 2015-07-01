package warrant

import "time"

type Group struct {
	ID          string
	DisplayName string
	Version     int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
