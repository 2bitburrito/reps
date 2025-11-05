package ghcloner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/2bitburrito/reps/cmd/reps/messages"
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

	cmd := exec.CommandContext(ctx, "git", "clone", choice[1])

	// Directly connect to stdout/stderr to avoid output capture issues
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run() // Use Run() instead of Start() to wait for completion

	// Always cancel the context to clean up the watchCtx goroutine
	if gc.ctxCancel != nil {
		cancel := *gc.ctxCancel
		cancel()
		gc.ctxCancel = nil
	}

	if err != nil {
		fmt.Printf("Git clone failed: %v\n", err)
		return err
	}
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
		gc.ctxCancel = nil
	}
}
