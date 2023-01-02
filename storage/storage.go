package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"go.uber.org/zap"
)

const (
	MetadataDirname = "metadata"
	SourceDirname   = "src"
	RepoFilename    = "repo.json"
)

type Storage struct {
	workingDir   string
	gcCandidates []string
	logger       *zap.SugaredLogger
}

func NewStorage(workingDir string) (*Storage, error) {
	s := &Storage{
		workingDir: workingDir,
		logger:     zap.S().Named("storage"),
	}
	if err := s.prepareGc(); err != nil {
		return nil, err
	}
	s.logger.Debug("initialized storage")
	return s, nil
}

func (s *Storage) RepoPath(prefix, baseUrl, owner, repo string) string {
	return filepath.Clean(filepath.Join(s.workingDir, prefix, baseUrl, owner, repo))
}

func (s *Storage) ListRepoPaths(prefix, baseUrl, owner string, regex *regexp.Regexp, glob string) ([]string, error) {
	var res []string
	root := filepath.Join(s.workingDir, prefix, baseUrl, owner)
	dirs, err := os.ReadDir(root)
	if err != nil {
		return []string{}, err
	}
	for _, dir := range dirs {
		if regex.Match([]byte(dir.Name())) {
			files, err := filepath.Glob(filepath.Join(root, dir.Name(), glob))
			if err != nil {
				return []string{}, err
			}
			res = append(res, files...)
		}
	}
	return res, nil
}

func (s *Storage) ListDirs() ([]string, error) {
	dirs, err := filepath.Glob(filepath.Join(s.workingDir, "*/*/*/*"))
	if err != nil {
		return []string{}, err
	}
	return dirs, nil
}

func (s *Storage) UpdateRepoMetadata(domain, name, repo string, data []byte) error {
	s.logger.Info(fmt.Sprintf("update repo metadata: %s/%s/%s", domain, name, repo))
	path := s.RepoPath(MetadataDirname, domain, name, repo)
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(path, RepoFilename), data, 0644); err != nil {
		return err
	}
	s.flagActive(path)
	return nil
}

func (s *Storage) UpdateSource(domain, name, repo, cloneUrl, tokenEnvvarName string) error {
	s.logger.Info(fmt.Sprintf("update source: %s/%s/%s", domain, name, repo))
	token := os.Getenv(tokenEnvvarName)
	path := s.RepoPath(SourceDirname, domain, name, repo)
	gitRepo, err := git.PlainOpen(path)
	if errors.Is(err, git.ErrRepositoryNotExists) {
		gitRepo, err = git.PlainClone(path, false, &git.CloneOptions{
			URL:   cloneUrl,
			Depth: 1,
			Auth:  &http.BasicAuth{Username: "username", Password: token},
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
			Auth:  &http.BasicAuth{Username: "username", Password: token},
		}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return err
		}
	}
	s.flagActive(path)
	return nil
}

func (s *Storage) prepareGc() error {
	dirs, err := s.ListDirs()
	if err != nil {
		return fmt.Errorf("failed preparing for gc: %s", err.Error())
	}
	s.gcCandidates = dirs
	return nil
}

func (s *Storage) flagActive(path string) {
	for i, v := range s.gcCandidates {
		if v == path {
			s.gcCandidates = append(s.gcCandidates[:i], s.gcCandidates[i+1:]...)
			break
		}
	}
}

func (s *Storage) DoGc() error {
	for _, path := range s.gcCandidates {
		s.logger.Info(fmt.Sprintf("gc: remove unused dir: %s", path))
		if err := os.RemoveAll(path); err != nil {
			return err
		}
		// TODO: remove parent directory if empty
	}
	return nil
}
