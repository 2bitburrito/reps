package fzf

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/2bitburrito/reps/cmd/reps/messages"
	"github.com/2bitburrito/reps/internal/common"
	"github.com/anthdm/hollywood/actor"
)

type fzfActor struct {
	id          string
	ActorEngine *actor.Engine
	PID         *actor.PID
	pipeReader  *io.PipeReader
	pipeWriter  *io.PipeWriter
	fzfCancel   context.CancelFunc // To cancel the fzf command context
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
		f.ActorEngine = ctx.Engine()
		f.PID = ctx.PID()
	case actor.Stopped:
		f.Finished()
	case messages.Initialise:
		go f.run(ctx)
	case messages.RepoMessage:
		f.pipeMessageToFfz(msg, ctx)
	case messages.FetchesComplete:
		if f.pipeWriter != nil {
			f.pipeWriter.Close()
			f.pipeWriter = nil
		}
	}
}

func (f *fzfActor) run(ctx *actor.Context) {
	fzfCtx, cancel := context.WithCancel(context.Background())
	f.fzfCancel = cancel

	fzfCmd := exec.CommandContext(fzfCtx, "fzf",
		"--with-nth=1", // only show the repo name in the list
		"--delimiter="+common.StrDelim,
		"--preview", "echo URL: {2}\n\necho Description: {3} ",
		"--style", "full",
		"--header", "Select a repo to clone.",
		"--tmux", "center,80%",
	)

	pipeReader, pipeWriter := io.Pipe()
	fzfCmd.Stdin = pipeReader
	f.pipeWriter = pipeWriter
	f.pipeReader = pipeReader

	stdout, err := fzfCmd.StdoutPipe()
	if err != nil {
		ctx.Engine().BroadcastEvent(messages.Failure{
			Source:  *ctx.PID(),
			Message: err.Error(),
		})
		return
	}

	err = fzfCmd.Start()
	if err != nil {
		ctx.Engine().BroadcastEvent(messages.Failure{
			Source:  *ctx.PID(),
			Message: err.Error(),
		})
		return
	}

	scanner := bufio.NewScanner(stdout)
	var choice string
	if scanner.Scan() {
		choice = scanner.Text()

		if f.pipeWriter != nil {
			f.pipeWriter.Close()
			f.pipeWriter = nil
		}
	}

	err = fzfCmd.Wait()

	if f.fzfCancel != nil {
		f.fzfCancel()
		f.fzfCancel = nil
	}

	if err != nil {
		// Check if it's a user cancellation exit code 130
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 130 || exitErr.ExitCode() == 1 {
				ctx.Engine().BroadcastEvent(messages.Shutdown{})
				return
			}
		}
		ctx.Engine().BroadcastEvent(messages.Failure{
			Source:  *ctx.PID(),
			Message: err.Error(),
		})
		return
	}

	choice = strings.TrimSpace(choice)
	if choice == "" {
		fmt.Println("No selection made.")
		ctx.Engine().BroadcastEvent(messages.Shutdown{})
		return
	}

	choiceSlice := strings.Split(choice, common.StrDelim)
	ctx.Send(ctx.Parent(), messages.FetchRepo{
		RepoChoice: choiceSlice,
	})

}

func (f *fzfActor) pipeMessageToFfz(msg messages.RepoMessage, ctx *actor.Context) {
	if f.pipeWriter == nil {
		return
	}
	strReader := common.FormatRepoList(msg.GetRepos())
	_, err := io.Copy(f.pipeWriter, strReader)
	if err != nil {
		errMsg := fmt.Sprintf("error while piping into fzf reader: %v", err)
		fmt.Println(errMsg)
		ctx.Engine().BroadcastEvent(messages.Failure{
			Source:  *ctx.PID(),
			Message: errMsg,
		})
		return
	}
}

func (f *fzfActor) Finished() {

	// Cancel fzf context if still running
	if f.fzfCancel != nil {
		f.fzfCancel()
		f.fzfCancel = nil
	}

	// Close pipes if still open
	if f.pipeWriter != nil {
		f.pipeWriter.Close()
		f.pipeWriter = nil
	}
	if f.pipeReader != nil {
		f.pipeReader.Close()
		f.pipeReader = nil
	}

	// Unsubscribe first to prevent deadletter buildup
	if f.ActorEngine != nil && f.PID != nil {
		f.ActorEngine.Unsubscribe(f.PID)
	}

	// make sure ActorEngine and PID are set
	if f.ActorEngine == nil {
		slog.Error("fzfActor.actorEngine is <nil>")
	}
	if f.PID == nil {
		slog.Error("fzfActor.PID is <nil>")
	}
}
