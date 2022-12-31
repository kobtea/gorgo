package check

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/kobtea/gorgo/config"
	"github.com/kobtea/gorgo/fetch"
	"github.com/open-policy-agent/conftest/output"
	"github.com/open-policy-agent/conftest/runner"
)

func Check(ctx context.Context, cfg *config.Config) error {
	var result []output.CheckResult
	for _, elm := range cfg.Users {
		pat := filepath.Join(cfg.WorkingDir, fetch.MetadataDirname, elm.Name, "*", fetch.RepoFilename)
		files, err := filepath.Glob(pat)
		if err != nil {
			return err
		}
		var matchFiles []string
		for _, file := range files {
			l := strings.Split(file, "/")
			repoName := l[len(l)-2]
			if elm.Regex.Match([]byte(repoName)) {
				matchFiles = append(matchFiles, file)
			}
		}
		r := runner.TestRunner{
			AllNamespaces: true,
			Policy:        elm.RepoPolicies,
		}
		res, err := r.Run(ctx, matchFiles)
		if err != nil {
			return err
		}
		result = append(result, res...)
	}
	// FIXME: support multi format
	outputter := output.Get("", output.Options{})
	if err := outputter.Output(result); err != nil {
		return err
	}
	return nil
}
