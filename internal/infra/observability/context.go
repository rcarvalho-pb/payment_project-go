package observability

import "context"

type correlationKey struct{}

func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationKey{}, id)
}

func CorrelationIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(correlationKey{}).(string)
	return id, ok
}
