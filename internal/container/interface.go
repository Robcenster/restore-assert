package container

import (
	"context"
)

// Provider рулит жизненным циклом СУБД в Docker.
type Provider interface {
	Start(ctx context.Context) error
	ExecuteRestore(ctx context.Context, hostFilePath string) error
	GetConnectionString() string
	Stop(ctx context.Context) error
}
