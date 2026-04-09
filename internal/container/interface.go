package container

import (
	"context"
)

type Provider interface {
	Start(ctx context.Context) error
	ExecuteRestore(ctx context.Context, hostFilePath string) error
	GetConnectionString() string
	Stop(ctx context.Context) error
}
