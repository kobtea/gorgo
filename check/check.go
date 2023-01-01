package check

import (
	"context"

	"github.com/kobtea/gorgo/config"
	"github.com/kobtea/gorgo/fetch"
	"github.com/kobtea/gorgo/storage"
	"github.com/open-policy-agent/conftest/output"
	"github.com/open-policy-agent/conftest/runner"
)

func Check(ctx context.Context, cfg *config.Config) error {
	var result []output.CheckResult
	st := storage.NewStorage(cfg.WorkingDir)
	// metadata
	for _, elm := range cfg.Users {
		files, err := st.ListUserRepoPaths(fetch.MetadataDirname, "github.com", elm.Name, elm.Regex.Regexp, fetch.RepoFilename)
		if err != nil {
			return err
		}

		r := runner.TestRunner{
			AllNamespaces: true,
			Policy:        elm.RepoPolicies,
		}
		res, err := r.Run(ctx, files)
		if err != nil {
			return err
		}
		result = append(result, res...)
	}
	// source
	for _, elm := range cfg.Users {
		for _, srcPolicy := range elm.SrcPolicies {
			paths, err := st.ListUserRepoPaths("src", "github.com", elm.Name, elm.Regex.Regexp, srcPolicy.Input)
			if err != nil {
				return err
			}
			r := runner.TestRunner{
				AllNamespaces: true,
				Combine:       srcPolicy.Combine,
				Policy:        srcPolicy.Policies,
			}
			res, err := r.Run(ctx, paths)
			if err != nil {
				return err
			}
			result = append(result, res...)
		}
	}

	// FIXME: support multi format
	outputter := output.Get("", output.Options{})
	if err := outputter.Output(result); err != nil {
		return err
	}
	return nil
}
