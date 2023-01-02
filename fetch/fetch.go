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
	for _, ghConfig := range cfg.GithubConfigs {
		userm := map[string][]*config.Regexp{}
		for _, userRepoConfig := range ghConfig.UserRepoConfigs {
			userm[userRepoConfig.Name] = append(userm[userRepoConfig.Name], userRepoConfig.Regex)
		}
		for user, regexes := range userm {
			err := fetchUserRepositories(ctx, user, regexes, cfg.WorkingDir, &githubOption{
				domain:          ghConfig.Domain(),
				baseUrl:         ghConfig.ApiEndpoint,
				uploadUrl:       ghConfig.UploadEndpoint,
				tokenEnvvarName: ghConfig.EnvvarName(),
			})
			if err != nil {
				return err
			}
		}
		// TODO: support org
	}
	return nil
}

type githubOption struct {
	domain          string
	baseUrl         string
	uploadUrl       string
	tokenEnvvarName string
}

func newClient(ctx context.Context, option *githubOption) (*github.Client, error) {
	if len(option.tokenEnvvarName) == 0 {
		return nil, fmt.Errorf("require github api token")
	}
	token := os.Getenv(option.tokenEnvvarName)
	if len(token) == 0 {
		return nil, fmt.Errorf("require %s env var", option.tokenEnvvarName)
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	if len(option.baseUrl) == 0 {
		return github.NewClient(tc), nil
	} else {
		return github.NewEnterpriseClient(option.baseUrl, option.uploadUrl, tc)
	}
}

func fetchUserRepositories(ctx context.Context, name string, regexes []*config.Regexp, outputDir string, ghOption *githubOption) error {
	st := storage.NewStorage(outputDir)
	cli, err := newClient(ctx, ghOption)
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
					// metadata
					j, err := json.Marshal(repo)
					if err != nil {
						return err
					}
					dir := st.UserRepoPath(MetadataDirname, ghOption.domain, name, *repo.Name)
					if err = os.MkdirAll(dir, 0755); err != nil {
						return err
					}
					if err = os.WriteFile(filepath.Join(dir, RepoFilename), j, 0644); err != nil {
						return err
					}

					// source
					srcPath := st.UserRepoPath(SourceDirname, ghOption.domain, name, *repo.Name)
					gitRepo, err := git.PlainOpen(srcPath)
					if errors.Is(err, git.ErrRepositoryNotExists) {
						gitRepo, err = git.PlainClone(srcPath, false, &git.CloneOptions{
							URL:   *repo.CloneURL,
							Depth: 1,
							Auth:  &http.BasicAuth{Username: "username", Password: os.Getenv(ghOption.tokenEnvvarName)},
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
							Auth:  &http.BasicAuth{Username: "username", Password: os.Getenv(ghOption.tokenEnvvarName)},
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
