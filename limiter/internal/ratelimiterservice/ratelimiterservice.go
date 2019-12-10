package ratelimiter

import (
	"fmt"
	"time"

	storage "github.com/shanehowearth/ratelimiter/limiter/integrations/repository"
)

// RateLimitService -
type RateLimitService struct {
	store    storage.Store
	limit    int
	timespan time.Duration
}

func NewRateLimitService(store storage.Store, limit *int, timespan *time.Duration) (*RateLimitService, error) {
	if store == nil || limit == nil || timespan == nil {
		return nil, fmt.Errorf("store, limit, and timespan are mandatory fields")
	}
	r := &RateLimitService{store: store, limit: *limit, timespan: *timespan}
	return r, nil
}

func (r *RateLimitService) CheckReachedLimit(ip string) (bool, time.Duration, error) {
	over, minWait, err := r.store.CreateAndCheck(ip, r.limit, time.Now(), r.timespan)

	return over, minWait, err
}
