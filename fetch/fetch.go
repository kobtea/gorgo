package fetch

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/go-github/v48/github"
	"github.com/kobtea/gorgo/config"
	"github.com/kobtea/gorgo/storage"
	"golang.org/x/oauth2"
)

func Fetch(ctx context.Context, cfg *config.Config) error {
	storage, err := storage.NewStorage(cfg.WorkingDir)
	if err != nil {
		return err
	}
	for _, ghConfig := range cfg.GithubConfigs {
		userm := map[string][]*config.Regexp{}
		for _, userRepoConfig := range ghConfig.UserRepoConfigs {
			userm[userRepoConfig.Name] = append(userm[userRepoConfig.Name], userRepoConfig.Regex)
		}
		for user, regexes := range userm {
			err := fetchUserRepositories(ctx, storage, user, regexes, &githubOption{
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
	storage.DoGc()
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

func fetchUserRepositories(ctx context.Context, storage *storage.Storage, name string, regexes []*config.Regexp, ghOption *githubOption) error {
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
					if r.UsedWithRepo {
						j, err := json.Marshal(repo)
						if err != nil {
							return err
						}
						if err := storage.UpdateRepoMetadata(ghOption.domain, name, *repo.Name, j); err != nil {
							return err
						}
					}
					// source
					if r.UsedWithSrc {
						if err := storage.UpdateSource(ghOption.domain, name, *repo.Name, *repo.CloneURL, ghOption.tokenEnvvarName); err != nil {
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
