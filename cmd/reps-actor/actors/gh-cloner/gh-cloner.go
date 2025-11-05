package ghcloner

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
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
		gc.Finished()
	case messages.FetchRepo:
		fetchCtx, fetchCancel := context.WithCancel(context.Background())
		gc.ctxCancel = &fetchCancel
		gc.fetchRepo(msg, fetchCtx)
		ctx.Engine().BroadcastEvent(messages.Shutdown{})
	}
}

func (gc *GhCloner) fetchRepo(msg messages.FetchRepo, ctx context.Context) error {
	choice := msg.RepoChoice
	fmt.Println("\nCloning Repo:", choice[0])

	cmd := exec.CommandContext(ctx, "git", "clone", choice[1])
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	return nil
}

func (gc *GhCloner) Finished() {
	// Unsubscribe first to prevent deadletter buildup
	if gc.ActorEngine != nil && gc.PID != nil {
		gc.ActorEngine.Unsubscribe(gc.PID)
	}

	// make sure ActorEngine and PID are set
	if gc.ActorEngine == nil {
		slog.Error("ghcloner.actorEngine is <nil>")
	}
	if gc.PID == nil {
		slog.Error("ghcloner.PID is <nil>")
	}

	if gc.ctxCancel != nil {
		cancel := *gc.ctxCancel
		cancel()
	}

	// poision itself
	gc.ActorEngine.Poison(gc.PID)
}
