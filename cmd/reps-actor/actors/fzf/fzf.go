package fzf

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/2bitburrito/reps/cmd/reps-actor/messages"
	"github.com/2bitburrito/reps/internal/common"
	"github.com/anthdm/hollywood/actor"
)

type fzfActor struct {
	id          string
	ActorEngine *actor.Engine
	PID         *actor.PID
	pipeWriter  io.Writer
}

func New() actor.Producer {
	return func() actor.Receiver {
		return &fzfActor{}
	}
}

func (f *fzfActor) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		ctx.Engine().Subscribe(ctx.PID())
		log.Println("fzfActor.Started", "id", f.id)
		f.ActorEngine = ctx.Engine()
		f.PID = ctx.PID()
	case actor.Stopped:
		// Clean up here
		log.Println("fzfActor.Stopped", f.id)
		f.Finished()
	case messages.Initialise:
		go f.run(ctx)
	case messages.RepoPayload:
		f.pipeMessageToFfz(msg, ctx)
	}
}
func (f *fzfActor) run(ctx *actor.Context) {
	fzfCmd := exec.Command(
		"fzf",
		"--with-nth=1", // only show the repo name in the list
		"--delimiter="+common.StrDelim,
		"--preview", "echo URL: {2}\n\necho Description: {3} ",
		"--style", "full",
		"--header", "Select a repo to clone! - If this is the first run it could take a while to fetch",
	)

	pipeReader, pipeWriter := io.Pipe()
	fzfCmd.Stdin = pipeReader
	f.pipeWriter = pipeWriter
	out, err := fzfCmd.CombinedOutput()
	if err != nil {
		ctx.Engine().BroadcastEvent(messages.Failure{
			Source:  *ctx.PID(),
			Message: err.Error(),
		})
		return
	}
	choice := strings.TrimSpace(string(out))
	if choice == "" {
		fmt.Println("No selection made.")
	}

	choiceSlice := strings.Split(choice, common.StrDelim)
	ctx.Send(ctx.Parent(), messages.FetchRepo{
		RepoChoice: choiceSlice,
	})

}

func (f *fzfActor) pipeMessageToFfz(msg messages.RepoPayload, ctx *actor.Context) {
	strReader := common.FormatRepoList(msg.Repos)
	_, err := io.Copy(f.pipeWriter, strReader)
	if err != nil {
		errMsg := fmt.Sprintf("error while piping into fzf reader: %v", err)
		fmt.Println(errMsg)
		ctx.Engine().BroadcastEvent(messages.Failure{
			Source:  *ctx.PID(),
			Message: errMsg,
		})
		// do i need to poison self here?
		return
	}
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
