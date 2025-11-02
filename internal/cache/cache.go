package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/2bitburrito/reps/internal/common"
)

type Cache struct {
	set   map[string]struct{}
	Repos []common.Repo
}

func NewCache() *Cache {
	c := &Cache{
		set:   map[string]struct{}{},
		Repos: []common.Repo{},
	}
	return c
}

func (c *Cache) GetCachedRepos(org string) ([]common.Repo, error) {
	dirPath, err := getCachePath()
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("couldn't read dir: %w", err)
	}

	haveCacheFile := false
	for _, file := range files {
		if strings.Contains(file.Name(), org) {
			haveCacheFile = true
		}
	}

	if !haveCacheFile {
		return nil, nil
	}
	path := filepath.Join(dirPath, org+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't read file: %w", err)
	}
	var cachedRepos []common.Repo

	if err := json.Unmarshal(data, &cachedRepos); err != nil {
		return nil, fmt.Errorf("couldn't unmarshal cache: %w", err)
	}

	c.Repos = cachedRepos[:]
	c.setCacheSet(cachedRepos)
	return cachedRepos, nil
}

func (c *Cache) SaveRepoToCache(ctx context.Context, org string, repos []common.Repo) error {
	dirPath, err := getCachePath()
	if err != nil {
		fmt.Println("couldn't get cache path: ", err)
		return err
	}
	path := filepath.Join(dirPath, org+".json")
	jsonData, err := json.Marshal(repos)
	if err != nil {
		fmt.Println("couldn't marshal repo slice to json:", err)
		return err
	}

	if err := os.WriteFile(path, jsonData, 0600); err != nil {
		fmt.Println("error while writing to :", path, err)
		return err
	}
	return nil
}

func getCachePath() (string, error) {
	// "~/.cache/reps/<org>.json"
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	dirPath := filepath.Join(home, ".cache", "reps")

	err = os.MkdirAll(dirPath, 0700)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("couldn't create file structure for caching...", err)
			return "", err
		} else {
			return "", err
		}
	}
	return dirPath, nil
}

func (c *Cache) setCacheSet(repos []common.Repo) {
	for _, repo := range repos {
		c.set[repo.Name] = struct{}{}
	}
}

func (c *Cache) CheckCacheSet(repos []common.Repo) []common.Repo {
	var newRepos []common.Repo

	for _, repo := range repos {
		_, exists := c.set[repo.Name]
		if !exists {
			newRepos = append(newRepos, repo)
		}
	}
	return newRepos
}
