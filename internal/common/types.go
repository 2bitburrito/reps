package common

const StrDelim = "\u00a0"

type Repo struct {
	Name        string `json:"name"`
	Url         string `json:"url"`
	Description string `json:"description"`
}

const (
	IDRoot                = "0"
	IDDeviceSupervisor    = "0"
	IDDeviceServiceWorker = "0"
)

const (
	ActorTypeRoot        = "root"
	ActorTypeFzfWorker   = "fzf-actor"
	ActorTypeListGetter  = "list-getter"
	ActorTypeCacheWorker = "cache-worker"
	ActorTypeFetchWorker = "fetch-worker"
	ActorTypeGhCloner    = "gh-cloner"
)
