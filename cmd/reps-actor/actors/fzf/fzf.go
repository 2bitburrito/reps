package fzf

import (
	"log"
	"log/slog"

	"github.com/2bitburrito/reps/cmd/reps-actor/messages"
	"github.com/anthdm/hollywood/actor"
)

type fzfActor struct {
	id          string
	ActorEngine *actor.Engine
	PID         *actor.PID
}

func New() actor.Producer {
	return func() actor.Receiver {
		return &fzfActor{}
	}
}

// TODO: fix all this copypasta
func (f *fzfActor) Receive(c *actor.Context) {
	switch msg := c.Message().(type) {
	case actor.Started:
		log.Println("fzfActor.Started", "id", f.id)
		f.ActorEngine = c.Engine()
		f.PID = c.PID()
	case actor.Stopped:
		// Clean up here
		log.Println("fzfActor.Stopped", f.id)
		f.Finished()
	case messages.RepoPayload:
		f.pipeMessageToFfz(msg)
	}
}

func (f *fzfActor) pipeMessageToFfz(msg messages.RepoPayload) {

}

func (f *fzfActor) Finished() {
	// make sure ActorEngine and PID are set
	if f.ActorEngine == nil {
		slog.Error("fzfActor.actorEngine is <nil>")
	}
	if f.PID == nil {
		slog.Error("fzfActor.PID is <nil>")
	}

	// poision itself
	f.ActorEngine.Poison(f.PID)
}
