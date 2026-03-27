package authctx

import "context"

type ctxKey struct{}

// WithUserID stores the authenticated user ID in the context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ctxKey{}, userID)
}

// UserIDFromContext retrieves the authenticated user ID from the context.
// Returns ("", false) if no user ID is present.
func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxKey{}).(string)
	return v, ok
}
