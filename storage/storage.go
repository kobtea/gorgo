package storage

import (
	"os"
	"path/filepath"
	"regexp"
)

type Storage struct {
	workingDir string
}

func NewStorage(workingDir string) *Storage {
	return &Storage{
		workingDir: workingDir,
	}
}

func (s *Storage) UserRepoPath(prefix, baseUrl, user, repo string) string {
	return filepath.Join(s.workingDir, prefix, baseUrl, user, repo)
}

func (s *Storage) ListUserRepoPaths(prefix, baseUrl, user string, regex *regexp.Regexp, glob string) ([]string, error) {
	var res []string
	root := filepath.Join(s.workingDir, prefix, baseUrl, user)
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
