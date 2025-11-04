package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/2bitburrito/reps/internal/common"
)

func GetReposFromGH(org string, ctx context.Context) ([]common.Repo, error) {
	fmt.Println("Getting all repos...")
	cmd := exec.CommandContext(ctx, "gh", "repo", "list", org, "--limit", "10000", "--json", "name,description,url", "--no-archived")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("failed to run gh command: %v: %v", err, string(out))
		return nil, err
	}

	var repos []common.Repo
	if err := json.Unmarshal(out, &repos); err != nil {
		return nil, fmt.Errorf("failed to unmarshal list result from gh: %v", err)
	}

	return repos, nil
}
