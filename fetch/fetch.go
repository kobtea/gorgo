package fetch

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/go-github/v48/github"
	"github.com/kobtea/gorgo/config"
	"golang.org/x/oauth2"
)

const (
	metadataDirname = "metadata"
	repoFilename    = "repo.json"
)

func Fetch(cfg *config.Config) error {
	ctx := context.Background()
	userm := map[string][]*config.Regexp{}
	for _, user := range cfg.Users {
		userm[user.Name] = append(userm[user.Name], user.Regex)
	}
	for user, regexes := range userm {
		err := fetchUserRepositories(ctx, user, regexes, cfg.WorkingDir)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewClient(ctx context.Context) (*github.Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if len(token) == 0 {
		return nil, fmt.Errorf("require GITHUB_TOKEN env var")
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc), nil
}

func fetchUserRepositories(ctx context.Context, name string, regexes []*config.Regexp, outputDir string) error {
	cli, err := NewClient(ctx)
	if err != nil {
		return err
	}
	opt := &github.RepositoryListOptions{}
	for {
		repos, resp, err := cli.Repositories.List(ctx, name, opt)
		if err != nil {
			return err
		}
		for _, repo := range repos {
			for _, r := range regexes {
				if r.Match([]byte(*repo.Name)) {
					j, err := json.Marshal(repo)
					if err != nil {
						return err
					}
					dir := filepath.Join(outputDir, metadataDirname, name, *repo.Name)
					if err = os.MkdirAll(dir, 0755); err != nil {
						return err
					}

					if err = os.WriteFile(filepath.Join(dir, repoFilename), j, 0644); err != nil {
						return err
					}
				}
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return nil
}
