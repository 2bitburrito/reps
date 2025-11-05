package main

import (
	"os"
	"os/signal"
	"syscall"

	root "github.com/2bitburrito/reps/cmd/reps-actor/actors"
	"github.com/2bitburrito/reps/internal/cli"
	"github.com/2bitburrito/reps/internal/common"
	"github.com/anthdm/hollywood/actor"
)

func main() {
	e, err := actor.NewEngine(actor.NewEngineConfig())
	if err != nil {
		panic(err)
	}

	args := os.Args[1:]
	org := cli.GetOrg(args)

	rootPID := e.Spawn(
		root.New(org),
		common.ActorTypeRoot, actor.WithID(common.IDRoot),
	)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-sigChan
	e.Poison(rootPID)
}
