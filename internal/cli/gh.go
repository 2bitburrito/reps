package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"

	"github.com/2bitburrito/reps/internal/common"
)

func GetReposFromGH(org string, ctx context.Context) ([]common.Repo, error) {
	fmt.Println("Getting all repos...")
	cmd := exec.CommandContext(ctx, "gh", "repo", "list", org, "--limit", "10000", "--json", "name,description,url", "--no-archived")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start gh command: %v", err)
	}
	out, err := io.ReadAll(stdout)
	if err != nil {
		errMsg := fmt.Errorf("failed to read gh command output: %v", err)
		return nil, errMsg
	}

	var repos []common.Repo
	if err := json.Unmarshal(out, &repos); err != nil {
		return nil, fmt.Errorf("failed to unmarshal list result from gh: %v", err)
	}

	return repos, nil
}
