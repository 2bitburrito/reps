package root

import (
	"fmt"
	"log"

	"github.com/2bitburrito/reps/cmd/reps-actor/actors/fzf"
	ghcloner "github.com/2bitburrito/reps/cmd/reps-actor/actors/gh-cloner"
	listgetter "github.com/2bitburrito/reps/cmd/reps-actor/actors/list-getter"
	"github.com/2bitburrito/reps/cmd/reps-actor/messages"
	"github.com/2bitburrito/reps/internal/cli"
	"github.com/2bitburrito/reps/internal/common"
	"github.com/anthdm/hollywood/actor"
)

type actorEngine struct {
	id            string
	Org           string
	ActorEngine   *actor.Engine
	PID           *actor.PID
	listGetterPID *actor.PID
	fzfActorPID   *actor.PID
	ghClonerPID   *actor.PID
}

func New(org string) actor.Producer {
	return func() actor.Receiver {
		return &actorEngine{Org: org}
	}
}

func (ae *actorEngine) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Initialized:
		log.Println("engine.init")
		err := ae.initialize()
		if err != nil {
			fmt.Printf("Incorrect binaries installed: %v\n", err)
			fmt.Println("Please ensure you have installed both: 'fzf' & 'gh'")
			fmt.Println("Then run 'gh auth login'.")
			ctx.Engine().BroadcastEvent(messages.Failure{
				Source:  *ctx.PID(),
				Message: "no dependencies",
			})
			return
		}
		ae.spawnWorkers(ctx)
	case actor.Started:
		ctx.Engine().Subscribe(ctx.PID())
		log.Println("root.started", "id", ae.id)

		ctx.Engine().BroadcastEvent(messages.Initialise{
			Org: ae.Org,
		})
	case actor.Stopped:
	case messages.RepoPayloadFromFetch:
		ctx.Send(ae.fzfActorPID, msg)
	case messages.RepoPayloadFromCache:
		ctx.Send(ae.fzfActorPID, msg)
	case messages.FetchRepo:
		ctx.Send(ae.ghClonerPID, msg)
	case messages.FetchesComplete:
		ctx.Send(ae.fzfActorPID, msg)
	case messages.Shutdown:
		ctx.Engine().Poison(ae.fzfActorPID)
		ctx.Engine().Poison(ae.ghClonerPID)
		ctx.Engine().Poison(ae.listGetterPID)

		ctx.Engine().Poison(ctx.PID())
	}
}
func (ae *actorEngine) initialize() error {
	err := cli.CheckInstalledBinaries()
	if err != nil {
		return err
	}
	return nil
}

func (ae *actorEngine) spawnWorkers(ctx *actor.Context) {
	ae.listGetterPID = ctx.SpawnChild(listgetter.New(), common.ActorTypeListGetter, actor.WithID(common.ActorTypeListGetter))
	ae.fzfActorPID = ctx.SpawnChild(fzf.New(), common.ActorTypeFzfWorker, actor.WithID(common.ActorTypeFzfWorker))
	ae.ghClonerPID = ctx.SpawnChild(ghcloner.New(), common.ActorTypeGhCloner, actor.WithID(common.ActorTypeGhCloner))
}
