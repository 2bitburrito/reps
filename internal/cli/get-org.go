package cli

import (
	"fmt"
	"os"
)

func GetOrg(args []string) string {
	envOrg := os.Getenv("DEFAULT_ORG")

	if len(args) == 0 && envOrg == "" {
		fmt.Println("Usage is: reps <organisation-name>")
		fmt.Println("Or you can set default org with `export DEFAULT_ORG=<organisation-name>")
		os.Exit(1)
	}
	if len(envOrg) > 0 {
		return envOrg
	}
	return args[0]
}
