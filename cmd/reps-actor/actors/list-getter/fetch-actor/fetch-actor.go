package fetchactor

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/2bitburrito/reps/cmd/reps-actor/messages"
	"github.com/2bitburrito/reps/internal/cli"
	"github.com/anthdm/hollywood/actor"
)

type fetchActor struct {
	id          string
	ActorEngine *actor.Engine
	PID         *actor.PID
	ctxCancel   context.CancelFunc
}

func New() actor.Producer {
	return func() actor.Receiver {
		return &fetchActor{}
	}
}

func (fa *fetchActor) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		ctx.Engine().Subscribe(ctx.PID())
		log.Println("fetchActor.Started", fa.id)
		fa.ActorEngine = ctx.Engine()
		fa.PID = ctx.PID()
	case actor.Stopped:
		fa.Finished()
	case messages.Initialise:
		// Start streaming repos asynchronously
		fa.initializeFetch(msg, ctx)
	case messages.FetchRepo:
		if fa.ctxCancel != nil {
			fmt.Println("cancelling gh fetch")
			fa.ctxCancel()
			fa.ctxCancel = nil
		}
		fa.Finished()
	case messages.Shutdown:
		if fa.ctxCancel != nil {
			log.Println("fetchActor: cancelling fetch due to shutdown")
			fa.ctxCancel()
			fa.ctxCancel = nil
		}
	}
}

func (fa *fetchActor) initializeFetch(msg messages.Initialise, ctx *actor.Context) {
	ctxWCancel, cancel := context.WithCancel(context.Background())
	fa.ctxCancel = cancel

	// Run in goroutine to avoid blocking the actor
	go func() {
		repos, err := cli.GetReposFromGH(msg.Org, ctxWCancel)
		if err != nil && err != context.Canceled {
			ctx.Engine().BroadcastEvent(messages.Failure{
				Source:  *ctx.PID(),
				Message: err.Error(),
			})
		}
		ctx.Send(ctx.Parent(), messages.CheckCache{
			Repos: repos,
		})
	}()
}

func (fa *fetchActor) Finished() {
	// Unsubscribe first to prevent deadletter buildup
	if fa.ActorEngine != nil && fa.PID != nil {
		fa.ActorEngine.Unsubscribe(fa.PID)
	}

	// make sure ActorEngine and PID are set
	if fa.ActorEngine == nil {
		slog.Error("fetchActor.actorEngine is <nil>")
	}
	if fa.PID == nil {
		slog.Error("fetchActor.PID is <nil>")
	}
	if fa.ctxCancel != nil {
		fa.ctxCancel()
		fa.ctxCancel = nil
	}
	// poision self
	fa.ActorEngine.Poison(fa.PID)
}
