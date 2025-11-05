package listgetter

import (
	"fmt"
	"log"

	cacheactor "github.com/2bitburrito/reps/cmd/reps-actor/actors/list-getter/cache-actor"
	fetchactor "github.com/2bitburrito/reps/cmd/reps-actor/actors/list-getter/fetch-actor"
	"github.com/2bitburrito/reps/cmd/reps-actor/messages"
	"github.com/2bitburrito/reps/internal/common"
	"github.com/anthdm/hollywood/actor"
)

type listGetterActor struct {
	ActorEngine   *actor.Engine
	PID           *actor.PID
	cacheActorPID *actor.PID
	fetchActorPID *actor.PID
	id            string
	count         int8
}

func New() actor.Producer {
	return func() actor.Receiver {
		return &listGetterActor{}
	}
}

func (lg *listGetterActor) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		ctx.Engine().Subscribe(ctx.PID())
		log.Println("listgetter.Started", "id", lg.id)
		lg.ActorEngine = ctx.Engine()
		lg.PID = ctx.PID()
		lg.spawnChildren(ctx)
	case actor.Stopped:
		// Clean up here
		lg.Finished()
	case messages.CheckCache:
		// forward on to cache:
		ctx.Send(lg.cacheActorPID, msg)
	case messages.RepoPayloadFromFetch:
		// pass up to root
		lg.receiveRepo(ctx, msg)
	case messages.RepoPayloadFromCache:
		// pass up to root
		lg.receiveRepo(ctx, msg)
	}
}
func (lg *listGetterActor) spawnChildren(ctx *actor.Context) {
	lg.cacheActorPID = ctx.SpawnChild(cacheactor.New(), common.ActorTypeCacheWorker, actor.WithID(common.ActorTypeCacheWorker))
	lg.fetchActorPID = ctx.SpawnChild(fetchactor.New(), common.ActorTypeFetchWorker, actor.WithID(common.ActorTypeFetchWorker))
}

func (lg *listGetterActor) receiveRepo(ctx *actor.Context, msg messages.RepoMessage) {
	ctx.Send(ctx.Parent(), msg)
	lg.count += 1
	if lg.count >= 2 {
		ctx.Send(ctx.Parent(), messages.FetchesComplete{})
		lg.poisonChildren(ctx)
	}
}

func (lg *listGetterActor) poisonChildren(ctx *actor.Context) {
	ctx.Engine().Poison(lg.cacheActorPID)
	ctx.Engine().Poison(lg.fetchActorPID)
}
func (lg *listGetterActor) Finished() {
	// Unsubscribe first to prevent deadletter buildup
	if lg.ActorEngine != nil && lg.PID != nil {
		lg.ActorEngine.Unsubscribe(lg.PID)
	}

	// make sure ActorEngine and PID are set
	if lg.ActorEngine == nil {
		fmt.Println("listgetter.actorEngine is <nil>")
	}
	if lg.PID == nil {
		fmt.Println("listgetter.PID is <nil>")
	}

	// poision itself
	lg.ActorEngine.Poison(lg.PID)
}
