package listgetter

import (
	"log"
	"log/slog"

	cacheactor "github.com/2bitburrito/reps/cmd/reps-actor/actors/list-getter/cache-actor"
	fetchactor "github.com/2bitburrito/reps/cmd/reps-actor/actors/list-getter/fetch-actor"
	"github.com/2bitburrito/reps/internal/common"
	"github.com/anthdm/hollywood/actor"
)

type listGetterActor struct {
	id            string
	ActorEngine   *actor.Engine
	PID           *actor.PID
	cacheActorPID *actor.PID
	fetchActorPID *actor.PID
}

func New() actor.Producer {
	return func() actor.Receiver {
		return &listGetterActor{}
	}
}

func (lg *listGetterActor) Receive(ctx *actor.Context) {
	switch ctx.Message().(type) {
	case actor.Started:
		ctx.Engine().Subscribe(ctx.PID())
		log.Println("listgetter.Started", "id", lg.id)
		lg.ActorEngine = ctx.Engine()
		lg.PID = ctx.PID()
		lg.spawnChildren(ctx)
	case actor.Stopped:
		slog.Info("listgetter.Stopped", "id", lg.id)
		// Clean up here
		lg.Finished()
	}
}
func (lg *listGetterActor) spawnChildren(ctx *actor.Context) {
	lg.cacheActorPID = ctx.SpawnChild(cacheactor.New(), common.ActorTypeCacheWorker, actor.WithID(common.ActorTypeCacheWorker))
	lg.fetchActorPID = ctx.SpawnChild(fetchactor.New(), common.ActorTypeFetchWorker, actor.WithID(common.ActorTypeFetchWorker))
}

func (lg *listGetterActor) Finished() {
	// make sure ActorEngine and PID are set
	if lg.ActorEngine == nil {
		slog.Error("listgetter.actorEngine is <nil>")
	}
	if lg.PID == nil {
		slog.Error("listgetter.PID is <nil>")
	}

	// poision itself
	lg.ActorEngine.Poison(lg.PID)
}
