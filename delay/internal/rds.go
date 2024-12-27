package internal

import "time"

type Rds interface {
	HGetAll(key string) (map[string]string, error)
	HDel(key string, field string) error
	HSet(key string, field string, value any) error
	RPush(key string, value any) error
	BLPop(timeout time.Duration, keys ...string) ([]string, error)
}

func SetRds(r Rds) {
	if rds == nil {
		rds = r
	}
}

var rds Rds = nil
