package conf

import (
	"context"
)

type HandleFunc func(ctx context.Context) error
