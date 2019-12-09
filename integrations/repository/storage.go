package storage

import (
	"net"
	"time"
)

// Store -
type Store interface {
	CreateAndCheck(ip net.Addr, limit int, timestamp time.Time, timespan time.Duration) (bool, error)
	ReachedMax(ip net.Addr, limit int, timespan time.Duration) (bool, error)
	Create(ip net.Addr, timestamp time.Time) error
}
