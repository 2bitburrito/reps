package cacheactor

import (
	"fmt"
	"log"
	"log/slog"

	"github.com/2bitburrito/reps/cmd/reps-actor/messages"
	"github.com/2bitburrito/reps/internal/cache"
	"github.com/anthdm/hollywood/actor"
)

type cacheActor struct {
	id          string
	cache       *cache.Cache
	ActorEngine *actor.Engine
	PID         *actor.PID
}

func New() actor.Producer {
	return func() actor.Receiver {
		return &cacheActor{}
	}
}

func (ca *cacheActor) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		ctx.Engine().Subscribe(ctx.PID())
		log.Println("cacheActor.Started", "id", ca.id)
		ca.ActorEngine = ctx.Engine()
		ca.PID = ctx.PID()
		ca.cache = cache.NewCache()
	case actor.Stopped:
		// Clean up here
		log.Println("cacheActor.Stopped", "id", ca.id)
		ca.Finished()
	case messages.Initialise:
		ca.Initialise(msg, ctx)
	}
}

func (ca *cacheActor) Initialise(msg messages.Initialise, ctx *actor.Context) {
	ctx.Engine().Subscribe(ctx.PID())
	repos, err := ca.cache.GetCachedRepos(msg.Org)
	if err != nil {
		fmt.Println("error trying to get cached repos for:", msg.Org, err)
		return
	}
	rootPID := ctx.Sender()
	if rootPID == nil {
		fmt.Println("couldn't get parentPID in cache actor")
		return
	}
	ctx.Send(rootPID, messages.RepoPayload{
		Org:   msg.Org,
		Repos: repos,
	})

}

func (ca *cacheActor) Finished() {
	// make sure ActorEngine and PID are set
	if ca.ActorEngine == nil {
		slog.Error("tradeExecutor.actorEngine is <nil>")
	}
	if ca.PID == nil {
		slog.Error("tradeExecutor.PID is <nil>")
	}

	// poision itself
	ca.ActorEngine.Poison(ca.PID)
}
