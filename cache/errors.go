package cache

import "errors"

var (
	ErrCacheMiss = errors.New("cache: key is missing")
)
