package flows

import "context"

type Flow interface {
	Start(ctx context.Context)
}
