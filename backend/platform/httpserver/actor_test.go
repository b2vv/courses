package httpserver

import (
	"context"
	"testing"
)

func TestActorFromContext(t *testing.T) {
	t.Parallel()

	want := Actor{UserID: "40000000-0000-4000-8000-000000000003", Role: "admin", EmailVerified: true}
	for _, test := range []struct {
		name string
		ctx  context.Context
		want Actor
		ok   bool
	}{
		{name: "absent", ctx: context.Background()},
		{name: "present", ctx: WithActor(context.Background(), want), want: want, ok: true},
		{name: "wrong type", ctx: context.WithValue(context.Background(), actorKey{}, "admin")},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, ok := ActorFromContext(test.ctx)
			if got != test.want || ok != test.ok {
				t.Fatalf("got (%+v, %v), want (%+v, %v)", got, ok, test.want, test.ok)
			}
		})
	}
}
