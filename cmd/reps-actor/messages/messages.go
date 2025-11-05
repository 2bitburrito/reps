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

type RepoMessage interface {
	GetRepos() []common.Repo
}

// Contains a payload of Repos from the local cache
type RepoPayloadFromCache struct {
	Repos []common.Repo
}

func (r RepoPayloadFromCache) GetRepos() []common.Repo {
	return r.Repos
}

// Contains a payload of Repos from gh
type RepoPayloadFromFetch struct {
	Repos []common.Repo
}

func (r RepoPayloadFromFetch) GetRepos() []common.Repo {
	return r.Repos
}

type FetchRepo struct {
	RepoChoice []string
}

type CheckCache struct {
	Repos []common.Repo
}

type FetchesComplete struct{}

type Shutdown struct{}

type Failure struct {
	Source  actor.PID
	Message string
}
