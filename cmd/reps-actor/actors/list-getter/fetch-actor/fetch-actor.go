package fetchactor

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/2bitburrito/reps/cmd/reps-actor/messages"
	"github.com/2bitburrito/reps/internal/cli"
	"github.com/2bitburrito/reps/internal/common"
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
		log.Println("fetchActor.Stopped", fa.id)
		fa.Finished()
	case messages.Initialise:
		repos, err := fa.initialize(msg)
		if err != nil {
			ctx.Engine().BroadcastEvent(messages.Failure{
				Source:  *ctx.PID(),
				Message: err.Error(),
			})
			return
		}
		ctx.Engine().BroadcastEvent(messages.RepoPayload{
			Org:   msg.Org,
			Repos: repos,
		})
	case messages.FetchRepo:
		if fa.ctxCancel != nil {
			fmt.Println("cancelling gh fetch")
			fa.ctxCancel()
			fa.ctxCancel = nil
		}
		fa.Finished()
	}
}
func (fa *fetchActor) initialize(msg messages.Initialise) ([]common.Repo, error) {
	ctxWCancel, cancel := context.WithCancel(context.Background())
	fa.ctxCancel = cancel
	return cli.GetReposFromGH(msg.Org, ctxWCancel)
}

func (fa *fetchActor) Finished() {
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
