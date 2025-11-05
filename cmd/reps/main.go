package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/2bitburrito/reps/internal/cache"
	"github.com/2bitburrito/reps/internal/common"
)

const delim = "\u00a0"

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Usage is: reps <organisation-name>")
		os.Exit(1)
	}
	org := args[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sigChan
		fmt.Println("Gracefullly shutting down")
		cancel()
		time.Sleep(3 * time.Second)
		os.Exit(0)
	}()

	if err := checkInstalledBinaries(); err != nil {
		fmt.Printf("Incorrect binaries installed: %v\n", err)
		fmt.Println("Please ensure you have installed both: 'fzf' & 'gh'")
		fmt.Println("Then run 'gh auth login'.")
		cancel()
		os.Exit(1)
	}

	cache := cache.NewCache()
	_, err := cache.GetCachedRepos(org)
	if err != nil {
		fmt.Println("error getting locally cached data: ", err)
	}

	fzfCtx, cancelFzf := context.WithCancel(context.Background())
	defer cancelFzf()

	fetchedReposFromGH := make(chan *strings.Reader)
	go getFreshReposFromGH(ctx, cancelFzf, fetchedReposFromGH, org, cache)

	fzfCmd := exec.CommandContext(
		fzfCtx,
		"fzf",
		"--with-nth=1", // only show the repo name in the list
		"--delimiter="+delim,
		"--preview", "echo URL: {2}\n\necho Description: {3} ",
		"--style", "full",
		"--header", "Select a repo to clone - If this is the first run it could take a while to fetch",
	)

	pipeReader, pipeWriter := io.Pipe()
	fzfCmd.Stdin = pipeReader

	defer pipeWriter.Close()
	// Handle the pipe of new info from the cache/fetch asynchronously
	go func() {
		// pipe cached repos to fzf
		if len(cache.Repos) != 0 {
			formattedRepoList := formatRepoList(cache.Repos)
			_, err = io.Copy(pipeWriter, formattedRepoList)
			if err != nil {
				fmt.Println("error copying in pipe:", err)
			}
		}
		select {
		case <-ctx.Done():
			return
		case reader := <-fetchedReposFromGH:
			_, err = io.Copy(pipeWriter, reader)
			if err != nil {
				fmt.Println("error copying in pipe:", err)
			}
			err := pipeWriter.Close()
			if err != nil {
				fmt.Println("error closing writer in pipe:", err)
			}
		}
	}()

	out, err := fzfCmd.CombinedOutput()
	if err != nil {
		if !strings.Contains(err.Error(), "130") {
			fmt.Printf("fzf error: %v \n %s\n", err, string(out))
		}
		cancel()
		sigChan <- syscall.SIGKILL
		return
	}

	choice := strings.TrimSpace(string(out))
	if choice == "" {
		fmt.Println("No selection made.")
	}

	strPrts := strings.Split(choice, delim)
	fmt.Println("\nCloning:", strPrts[0])

	ghCmd := exec.CommandContext(ctx, "git", "clone", strPrts[1])
	out, err = ghCmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		fmt.Printf("failed to run git clone: %v: %s\n", err, out)
		cancel()
		return
	}
	fmt.Println(string(out))
}

func checkInstalledBinaries() error {
	cmd := exec.Command("gh", "version")
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("fzf", "--version")
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func getFreshReposFromGH(ctx context.Context, cancelFzf context.CancelFunc, ch chan *strings.Reader, org string, cache *cache.Cache) {
	defer close(ch)
	fmt.Println("Getting all repos...")
	// if this fails how do i kill the fzf process?
	// is passing the fzf's cancel func enough?
	cmd := exec.CommandContext(ctx, "gh", "repo", "list", org, "--limit", "10000", "--json", "name,description,url", "--no-archived")
	out, err := cmd.CombinedOutput()
	if err != nil {
		cancelFzf()
		fmt.Printf("failed to run gh command: %v: %v", err, string(out))
		return
	}

	var repos []common.Repo
	if err := json.Unmarshal(out, &repos); err != nil {
		log.Fatalf("failed to unmarshal list result from gh: %v", err)
	}
	newRepos := cache.CheckCacheSet(repos)
	go func() {
		if err = cache.SaveRepoToCache(org, repos); err != nil {
			fmt.Println("error caching new repo list: ", err)
		}
	}()
	ch <- formatRepoList(newRepos)
}

// Format Repo List prepares fzf input using a no break space delimiter
func formatRepoList(repos []common.Repo) *strings.Reader {
	var repoList []string
	for _, r := range repos {
		repoList = append(repoList, fmt.Sprintf("%s%s%s%s%s", r.Name, delim, r.Url, delim, r.Description))
	}
	reader := strings.NewReader(strings.Join(repoList, "\n"))
	return reader
}
