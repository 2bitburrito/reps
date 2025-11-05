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
	org         string
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
		ca.Finished()
	case messages.Initialise:
		ca.org = msg.Org
		ca.Initialise(msg, ctx)
	case messages.CheckCache:
		freshRepos := ca.cache.CheckCacheSet(msg.Repos)
		ctx.Send(ctx.Parent(), messages.RepoPayloadFromFetch{
			Repos: freshRepos,
		})
		err := ca.cache.SaveRepoToCache(ca.org, msg.Repos)
		if err != nil {
			fmt.Println("error saving repo to cache:", err)

		}
	}
}

func (ca *cacheActor) Initialise(msg messages.Initialise, ctx *actor.Context) {
	repos, err := ca.cache.GetCachedRepos(ca.org)
	if err != nil {
		fmt.Println("error trying to get cached repos for:", msg.Org, err)
		return
	}
	ctx.Send(ctx.Parent(), messages.RepoPayloadFromCache{
		Repos: repos,
	})
}

func (ca *cacheActor) Finished() {
	// Unsubscribe first to prevent deadletter buildup
	if ca.ActorEngine != nil && ca.PID != nil {
		ca.ActorEngine.Unsubscribe(ca.PID)
	}

	// make sure ActorEngine and PID are set
	if ca.ActorEngine == nil {
		slog.Error("cacheActor.actorEngine is <nil>")
	}
	if ca.PID == nil {
		slog.Error("cacheActor.PID is <nil>")
	}

	// poision itself
	ca.ActorEngine.Poison(ca.PID)
}
