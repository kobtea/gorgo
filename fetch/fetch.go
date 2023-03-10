package fetch

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/go-github/v48/github"
	"github.com/kobtea/gorgo/config"
	"github.com/kobtea/gorgo/log"
	"github.com/kobtea/gorgo/storage"
	"golang.org/x/oauth2"
)

type criteria struct {
	regexp   *config.Regexp
	archived bool
}

func Fetch(ctx context.Context, cfg *config.Config) error {
	log.GetLogger().Named("fetch").Info("fetch data")
	storage, err := storage.NewStorage(cfg.WorkingDir)
	if err != nil {
		return err
	}
	for _, ghConfig := range cfg.GithubConfigs {
		ghOpt := &githubOption{
			domain:          ghConfig.Domain(),
			baseUrl:         ghConfig.ApiEndpoint,
			uploadUrl:       ghConfig.UploadEndpoint,
			tokenEnvvarName: ghConfig.EnvvarName(),
		}
		ghCli, err := newClient(ctx, ghOpt)
		if err != nil {
			return err
		}
		ownerm := map[string][]criteria{}
		for _, repoConfig := range ghConfig.RepoConfigs {
			ownerm[repoConfig.Owner] = append(ownerm[repoConfig.Owner], criteria{repoConfig.Regex, repoConfig.Archived})
		}
		for owner, criterias := range ownerm {
			user, _, err := ghCli.Users.Get(ctx, owner)
			if err != nil {
				return err
			}
			if *user.Type == "User" {
				err := fetchUserRepositories(ctx, storage, owner, criterias, ghCli, ghOpt)
				if err != nil {
					return err
				}
			} else if *user.Type == "Organization" {
				err := fetchOrgRepositories(ctx, storage, owner, criterias, ghCli, ghOpt)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("un-supported user type: %s", *user.Type)
			}
		}
	}
	if err := storage.DoGc(); err != nil {
		return err
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

func fetchUserRepositories(ctx context.Context, storage *storage.Storage, name string, criterias []criteria, ghClient *github.Client, ghOption *githubOption) error {
	logger := log.GetLogger().Named("fetch.user")
	opt := &github.RepositoryListOptions{}
	for {
		repos, resp, err := ghClient.Repositories.List(ctx, name, opt)
		if err != nil {
			return err
		}
		for _, repo := range repos {
			for _, cri := range criterias {
				if cri.regexp.Match([]byte(*repo.Name)) && !(!cri.archived && *repo.Archived) {
					// metadata
					if cri.regexp.UsedWithRepo {
						j, err := json.Marshal(repo)
						if err != nil {
							logger.Warnf("failed to marshal metadata: %s/%s/%s: %s", ghOption.domain, name, *repo.Name, err.Error())
						} else {
							if err := storage.UpdateRepoMetadata(ghOption.domain, name, *repo.Name, j); err != nil {
								logger.Warnf("failed to update metadata: %s/%s/%s: %s", ghOption.domain, name, *repo.Name, err.Error())
							}
						}
					}
					// source
					if cri.regexp.UsedWithSrc {
						if err := storage.UpdateSource(ghOption.domain, name, *repo.Name, *repo.CloneURL, ghOption.tokenEnvvarName); err != nil {
							logger.Warnf("failed to update source: %s/%s/%s: %s", ghOption.domain, name, *repo.Name, err.Error())
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

func fetchOrgRepositories(ctx context.Context, storage *storage.Storage, name string, criterias []criteria, ghClient *github.Client, ghOption *githubOption) error {
	logger := log.GetLogger().Named("fetch.org")
	opt := &github.RepositoryListByOrgOptions{}
	for {
		repos, resp, err := ghClient.Repositories.ListByOrg(ctx, name, opt)
		if err != nil {
			return err
		}
		for _, repo := range repos {
			for _, cri := range criterias {
				if cri.regexp.Match([]byte(*repo.Name)) && !(!cri.archived && *repo.Archived) {
					// metadata
					if cri.regexp.UsedWithRepo {
						j, err := json.Marshal(repo)
						if err != nil {
							logger.Warnf("failed to marshal metadata: %s/%s/%s: %s", ghOption.domain, name, *repo.Name, err.Error())
						} else {
							if err := storage.UpdateRepoMetadata(ghOption.domain, name, *repo.Name, j); err != nil {
								logger.Warnf("failed to update metadata: %s/%s/%s: %s", ghOption.domain, name, *repo.Name, err.Error())
							}
						}
					}
					// source
					if cri.regexp.UsedWithSrc {
						if err := storage.UpdateSource(ghOption.domain, name, *repo.Name, *repo.CloneURL, ghOption.tokenEnvvarName); err != nil {
							logger.Warnf("failed to update source: %s/%s/%s: %s", ghOption.domain, name, *repo.Name, err.Error())
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
