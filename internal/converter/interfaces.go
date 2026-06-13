package converter

import (
	"context"
)

// Command executes an external command.
type Command interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}
