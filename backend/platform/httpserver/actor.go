package httpserver

import "context"

type actorKey struct{}

type Actor struct {
	UserID        string
	Role          string
	EmailVerified bool
}

func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, actorKey{}, actor)
}

func ActorFromContext(ctx context.Context) (Actor, bool) {
	actor, ok := ctx.Value(actorKey{}).(Actor)
	return actor, ok
}
