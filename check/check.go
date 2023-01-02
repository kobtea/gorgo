package check

import (
	"context"
	"fmt"

	"github.com/kobtea/gorgo/config"
	"github.com/kobtea/gorgo/fetch"
	"github.com/kobtea/gorgo/storage"
	"github.com/open-policy-agent/conftest/output"
	"github.com/open-policy-agent/conftest/runner"
)

func Check(ctx context.Context, cfg *config.Config) error {
	var result []output.CheckResult
	st := storage.NewStorage(cfg.WorkingDir)
	for _, ghConfig := range cfg.GithubConfigs {
		for _, userRepoConfig := range ghConfig.UserRepoConfigs {
			for _, ConftestConfig := range userRepoConfig.ConftestConfigs {
				var prefix string
				var glob string
				if ConftestConfig.Target == config.TargetRepo {
					prefix = fetch.MetadataDirname
					glob = fetch.RepoFilename
				} else if ConftestConfig.Target == config.TargetSrc {
					prefix = fetch.SourceDirname
					glob = ConftestConfig.Input
				} else {
					return fmt.Errorf("invalid target type: %s", ConftestConfig.Target)
				}

				files, err := st.ListUserRepoPaths(prefix, ghConfig.Domain(), userRepoConfig.Name, userRepoConfig.Regex.Regexp, glob)
				if err != nil {
					return err
				}
				r := runner.TestRunner{
					AllNamespaces: true,
					Combine:       ConftestConfig.Combine,
					Policy:        ConftestConfig.Policies,
				}
				res, err := r.Run(ctx, files)
				if err != nil {
					return err
				}
				result = append(result, res...)
			}
		}
	}

	// FIXME: support multi format
	outputter := output.Get("", output.Options{})
	if err := outputter.Output(result); err != nil {
		return err
	}
	return nil
}
