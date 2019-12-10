package storage

import (
	"time"
)

// Store -
type Store interface {
	CreateAndCheck(ip string, limit int, timestamp time.Time, timespan time.Duration) (bool, time.Duration, error)
	ReachedMax(ip string, limit int, timespan time.Duration) (bool, time.Duration, error)
	Create(ip string, timestamp time.Time) error
}
