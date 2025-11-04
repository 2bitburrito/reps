package common

import (
	"fmt"
	"strings"
)

// Format Repo List prepares fzf input using a no break space delimiter
func FormatRepoList(repos []Repo) *strings.Reader {
	var repoList []string
	for _, r := range repos {
		repoList = append(repoList, fmt.Sprintf("%s%s%s%s%s", r.Name, StrDelim, r.Url, StrDelim, r.Description))
	}
	reader := strings.NewReader(strings.Join(repoList, "\n"))
	return reader
}

func FormatChoiceOutput(out []byte) []string {
	choice := strings.TrimSpace(string(out))
	if choice == "" {
		fmt.Println("No selection made.")
		return nil
	}

	strPrts := strings.Split(choice, StrDelim)
	return strPrts
}
