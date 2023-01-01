package fetch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v48/github"
	"github.com/kobtea/gorgo/config"
	"github.com/kobtea/gorgo/storage"
	"golang.org/x/oauth2"
)

const (
	MetadataDirname = "metadata"
	SourceDirname   = "src"
	RepoFilename    = "repo.json"
)

func Fetch(ctx context.Context, cfg *config.Config) error {
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

func newClient(ctx context.Context) (*github.Client, error) {
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
	st := storage.NewStorage(outputDir)
	cli, err := newClient(ctx)
	if err != nil {
		return err
	}
	opt := &github.RepositoryListOptions{}
	for {
		// TODO: support ghe domain
		repos, resp, err := cli.Repositories.List(ctx, name, opt)
		if err != nil {
			return err
		}
		for _, repo := range repos {
			for _, r := range regexes {
				if r.Match([]byte(*repo.Name)) {
					// metadata
					j, err := json.Marshal(repo)
					if err != nil {
						return err
					}
					dir := st.UserRepoPath(MetadataDirname, "github.com", name, *repo.Name)
					if err = os.MkdirAll(dir, 0755); err != nil {
						return err
					}
					if err = os.WriteFile(filepath.Join(dir, RepoFilename), j, 0644); err != nil {
						return err
					}

					// source
					srcPath := st.UserRepoPath(SourceDirname, "github.com", name, *repo.Name)
					gitRepo, err := git.PlainOpen(srcPath)
					if errors.Is(err, git.ErrRepositoryNotExists) {
						gitRepo, err = git.PlainClone(srcPath, false, &git.CloneOptions{
							URL:   *repo.CloneURL,
							Depth: 1,
							Auth:  &http.BasicAuth{Username: "username", Password: os.Getenv("GITHUB_TOKEN")},
						})
						if err != nil {
							return err
						}
					} else if err != nil {
						return err
					} else {
						wt, err := gitRepo.Worktree()
						if err != nil {
							return err
						}
						if err = wt.Pull(&git.PullOptions{
							Depth: 1,
							Auth:  &http.BasicAuth{Username: "username", Password: os.Getenv("GITHUB_TOKEN")},
						}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
							return err
						}
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
