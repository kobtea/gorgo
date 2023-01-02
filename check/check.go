package check

import (
	"context"
	"fmt"

	"github.com/kobtea/gorgo/config"
	"github.com/kobtea/gorgo/storage"
	"github.com/open-policy-agent/conftest/output"
	"github.com/open-policy-agent/conftest/runner"
	"go.uber.org/zap"
)

func Check(ctx context.Context, cfg *config.Config) error {
	logger := zap.S().Named("check")
	logger.Info("check data")
	var result []output.CheckResult
	st, err := storage.NewStorage(cfg.WorkingDir)
	if err != nil {
		return err
	}
	for _, ghConfig := range cfg.GithubConfigs {
		for _, userRepoConfig := range ghConfig.UserRepoConfigs {
			for _, ConftestConfig := range userRepoConfig.ConftestConfigs {
				var prefix string
				var glob string
				if ConftestConfig.Target == config.TargetRepo {
					prefix = storage.MetadataDirname
					glob = storage.RepoFilename
				} else if ConftestConfig.Target == config.TargetSrc {
					prefix = storage.SourceDirname
					glob = ConftestConfig.Input
				} else {
					return fmt.Errorf("invalid target type: %s", ConftestConfig.Target)
				}

				files, err := st.ListRepoPaths(prefix, ghConfig.Domain(), userRepoConfig.Name, userRepoConfig.Regex.Regexp, glob)
				if err != nil {
					return err
				}
				r := runner.TestRunner{
					AllNamespaces: true,
					Combine:       ConftestConfig.Combine,
					Policy:        ConftestConfig.Policies,
				}
				logger.Debugw(fmt.Sprintf("conftest"), "policy", ConftestConfig.Policies, "input", files)
				res, err := r.Run(ctx, files)
				if err != nil {
					return err
				}
				result = append(result, res...)
			}
		}
		for _, orgRepoConfig := range ghConfig.OrgRepoConfigs {
			for _, ConftestConfig := range orgRepoConfig.ConftestConfigs {
				var prefix string
				var glob string
				if ConftestConfig.Target == config.TargetRepo {
					prefix = storage.MetadataDirname
					glob = storage.RepoFilename
				} else if ConftestConfig.Target == config.TargetSrc {
					prefix = storage.SourceDirname
					glob = ConftestConfig.Input
				} else {
					return fmt.Errorf("invalid target type: %s", ConftestConfig.Target)
				}

				files, err := st.ListRepoPaths(prefix, ghConfig.Domain(), orgRepoConfig.Name, orgRepoConfig.Regex.Regexp, glob)
				if err != nil {
					return err
				}
				r := runner.TestRunner{
					AllNamespaces: true,
					Combine:       ConftestConfig.Combine,
					Policy:        ConftestConfig.Policies,
				}
				logger.Debugw(fmt.Sprintf("conftest"), "policy", ConftestConfig.Policies, "input", files)
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
