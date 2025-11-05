package root

import (
	"fmt"
	"os"

	"github.com/2bitburrito/reps/cmd/reps/actors/fzf"
	ghcloner "github.com/2bitburrito/reps/cmd/reps/actors/gh-cloner"
	listgetter "github.com/2bitburrito/reps/cmd/reps/actors/list-getter"
	"github.com/2bitburrito/reps/cmd/reps/messages"
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

		ctx.Engine().BroadcastEvent(messages.Initialise{
			Org: ae.Org,
		})
	case actor.Stopped:
	case messages.RepoPayloadFromFetch:
		ctx.Send(ae.fzfActorPID, msg)
	case messages.RepoPayloadFromCache:
		ctx.Send(ae.fzfActorPID, msg)
	case messages.FetchRepo:
		// Send to listgetter first to cancel any ongoing fetch
		ctx.Send(ae.listGetterPID, msg)
		// Then send to ghcloner to start the clone
		ctx.Send(ae.ghClonerPID, msg)
	case messages.FetchesComplete:
		ctx.Send(ae.fzfActorPID, msg)
	case messages.Failure:
		ae.poisonTree(ctx)
		fmt.Println(msg.Message)
		os.Exit(0)
	case messages.Shutdown:
		ae.poisonTree(ctx)
		os.Exit(0)
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

func (ae *actorEngine) poisonTree(ctx *actor.Context) {
	ctx.Engine().Poison(ae.fzfActorPID)
	ctx.Engine().Poison(ae.ghClonerPID)
	ctx.Engine().Poison(ae.listGetterPID)

	ctx.Engine().Poison(ctx.PID())
}
