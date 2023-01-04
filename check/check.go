package check

import (
	"context"
	"fmt"

	"github.com/kobtea/gorgo/config"
	"github.com/kobtea/gorgo/log"
	"github.com/kobtea/gorgo/storage"
	"github.com/open-policy-agent/conftest/output"
	"github.com/open-policy-agent/conftest/runner"
)

func Check(ctx context.Context, cfg *config.Config, out string) error {
	logger := log.GetLogger().Named("check")
	defer log.MustSync()
	logger.Info("check data")
	var result []output.CheckResult
	st, err := storage.NewStorage(cfg.WorkingDir)
	if err != nil {
		return err
	}
	for _, ghConfig := range cfg.GithubConfigs {
		for _, repoConfig := range ghConfig.RepoConfigs {
			for _, ConftestConfig := range repoConfig.ConftestConfigs {
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

				files, err := st.ListRepoPaths(prefix, ghConfig.Domain(), repoConfig.Owner, repoConfig.Regex.Regexp, glob)
				if err != nil {
					return err
				}
				r := runner.TestRunner{
					AllNamespaces: true,
					Combine:       ConftestConfig.Combine,
					Policy:        ConftestConfig.Policies,
				}
				logger.Debugw("conftest", "policy", ConftestConfig.Policies, "input", files)
				res, err := r.Run(ctx, files)
				if err != nil {
					return err
				}
				result = append(result, res...)
			}
		}
	}

	outputter := output.Get(out, output.Options{})
	if err := outputter.Output(result); err != nil {
		return err
	}
	return nil
}
