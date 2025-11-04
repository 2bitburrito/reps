package messages

import (
	"github.com/2bitburrito/reps/internal/common"
	"github.com/anthdm/hollywood/actor"
)

// Initialise is a broadcast event which is sent to
// all actors on init
type Initialise struct {
	Org string
}

// Contains a payload of Repos either from the local
// cache or directly from GH cli
type RepoPayload struct {
	Org   string
	Repos []common.Repo
}

type FetchRepo struct {
	RepoChoice []string
}

type Failure struct {
	Source  actor.PID
	Message string
}
