package ghcloner

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os/exec"

	"github.com/2bitburrito/reps/cmd/reps-actor/messages"
	"github.com/anthdm/hollywood/actor"
)

type GhCloner struct {
	id          string
	ActorEngine *actor.Engine
	PID         *actor.PID
	ctxCancel   *context.CancelFunc
}

func New() actor.Producer {
	return func() actor.Receiver {
		return &GhCloner{}
	}
}

func (gc *GhCloner) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		ctx.Engine().Subscribe(ctx.PID())
		log.Println("ghcloner.Started", "id", gc.id)
		gc.ActorEngine = ctx.Engine()
		gc.PID = ctx.PID()
	case actor.Stopped:
		// Clean up here
		log.Println("ghcloner.Stopped", "id", gc.id)
		gc.Finished()
	case messages.FetchRepo:
		fetchCtx, fetchCancel := context.WithCancel(context.Background())
		gc.ctxCancel = &fetchCancel
		gc.fetchRepo(msg, fetchCtx)
	}
}

func (gc *GhCloner) fetchRepo(msg messages.FetchRepo, ctx context.Context) {
	choice := msg.RepoChoice
	fmt.Println("\nCloning Repo:", choice[0])

	ghCmd := exec.CommandContext(ctx, "git", "clone", choice[1])
	// TODO: Figure out how to pipe output for multi line outputs:
	out, err := ghCmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		fmt.Printf("failed to run git clone: %v: %s\n", err, out)
		return
	}
	fmt.Println(string(out))
}

func (gc *GhCloner) Finished() {
	// make sure ActorEngine and PID are set
	if gc.ActorEngine == nil {
		slog.Error("ghcloner.actorEngine is <nil>")
	}
	if gc.PID == nil {
		slog.Error("ghcloner.PID is <nil>")
	}
	cancel := *gc.ctxCancel
	cancel()

	// poision itself
	gc.ActorEngine.Poison(gc.PID)
}
