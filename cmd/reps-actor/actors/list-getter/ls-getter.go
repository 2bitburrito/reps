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
		log.Println("listgetter.Started", "id", lg.id)

		// set actorEngine and PID
		lg.ActorEngine = ctx.Engine()
		lg.PID = ctx.PID()

		lg.spawnChildren(ctx)

		// subscribe to price updates
		// lg.ActorEngine.Send(lg.priceWatcherPID, types.Subscribe{Sendto: lg.PID})

	case actor.Stopped:
		slog.Info("tradeExecutor.Stopped", "id", lg.id)
		// Clean up here

		// case types.PriceUpdate:
		// 	// update the price
		// 	lg.processUpdate(msg)
		//
		// case types.TradeInfoRequest:
		// 	slog.Info("tradeExecutor.TradeInfoRequest", "id", lg.id, "wallet", lg.wallet)
		//
		// 	// handle the request
		// 	lg.handleTradeInfoRequest(c)

		// case types.CancelOrderRequest:
		// 	slog.Info("tradeExecutor.CancelOrderRequest", "id", lg.id, "wallet", lg.wallet)
		//
		// 	// update status
		// 	lg.status = "cancelled"
		//
		// 	// stop the executor
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
		slog.Error("tradeExecutor.actorEngine is <nil>")
	}
	if lg.PID == nil {
		slog.Error("tradeExecutor.PID is <nil>")
	}

	// // unsubscribe from price updates
	// lg.ActorEngine.Send(lg.priceWatcherPID, types.Unsubscribe{Sendto: lg.PID})

	// poision itself
	lg.ActorEngine.Poison(lg.PID)
}
