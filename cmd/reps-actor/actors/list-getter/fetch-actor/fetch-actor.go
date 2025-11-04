package fetchactor

import (
	"log"
	"log/slog"

	"github.com/anthdm/hollywood/actor"
)

type fetchActor struct {
	id          string
	ActorEngine *actor.Engine
	PID         *actor.PID
}

func New() actor.Producer {
	return func() actor.Receiver {
		return &fetchActor{}
	}
}

// TODO: fix all this copypasta
func (fa *fetchActor) Receive(c *actor.Context) {
	switch c.Message().(type) {
	case actor.Started:
		log.Println("fetchActor.Started", "id", fa.id)

		// set actorEngine and PID
		fa.ActorEngine = c.Engine()
		fa.PID = c.PID()

		// subscribe to price updates
		// f.ActorEngine.Send(f.priceWatcherPID, types.Subscribe{Sendto: f.PID})

	case actor.Stopped:
		log.Println("tradeExecutor.Stopped", "id", fa.id)
		// Clean up here

		// case types.PriceUpdate:
		// 	// update the price
		// 	f.processUpdate(msg)
		//
		// case types.TradeInfoRequest:
		// 	slog.Info("tradeExecutor.TradeInfoRequest", "id", f.id, "wallet", f.wallet)
		//
		// 	// handle the request
		// 	f.handleTradeInfoRequest(c)

		// case types.CancelOrderRequest:
		// 	slog.Info("tradeExecutor.CancelOrderRequest", "id", f.id, "wallet", f.wallet)
		//
		// 	// update status
		// 	f.status = "cancelled"
		//
		// 	// stop the executor
		fa.Finished()
	}
}
func (fa *fetchActor) Finished() {
	// make sure ActorEngine and PID are set
	if fa.ActorEngine == nil {
		slog.Error("tradeExecutor.actorEngine is <nil>")
	}
	if fa.PID == nil {
		slog.Error("tradeExecutor.PID is <nil>")
	}

	// // unsubscribe from price updates
	// f.ActorEngine.Send(f.priceWatcherPID, types.Unsubscribe{Sendto: f.PID})

	// poision itself
	fa.ActorEngine.Poison(fa.PID)
}
