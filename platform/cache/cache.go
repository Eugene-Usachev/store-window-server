package cache

import (
	"context"
)

type Cache interface {
	IsNegativeCase(ctx context.Context, table string, key string) bool

	GetString(ctx context.Context, table string, key string) (res string, isExist bool)
	GetBytes(ctx context.Context, table string, key string) (res []byte, isExist bool)

	SetNegativeCase(ctx context.Context, table string, key string)
	SetString(ctx context.Context, table string, key string, value string)
	SetBytes(ctx context.Context, table string, key string, value []byte)

	Delete(ctx context.Context, table string, key string) error
}

func MustCreateMainCache() Cache {
	return MustCreateRedisCache()
}
