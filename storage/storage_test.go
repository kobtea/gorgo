package storage

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorage_ListUserRepoPaths(t *testing.T) {
	s := Storage{workingDir: "../testdata"}
	res, err := s.ListUserRepoPaths("metadata", "github.com", "kobtea", regexp.MustCompile(`foo`), "")
	assert.NoError(t, err)
	assert.ElementsMatch(t, res, []string{
		"../testdata/metadata/github.com/kobtea/foo_one",
		"../testdata/metadata/github.com/kobtea/foo_two",
	})
}
