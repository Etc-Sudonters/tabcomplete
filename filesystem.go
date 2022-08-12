package tabcomplete

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/sahilm/fuzzy"
)

type FileSystemTabCompletion struct {
	pathSep       string
	IncludeHidden bool
}

func NewFileSystemTabCompletion() FileSystemTabCompletion {
	return FileSystemTabCompletion{
		pathSep: string(os.PathSeparator),
	}
}

func mapf[A, B any](as []A, f func(A) B) []B {
	bees := make([]B, len(as))

	for i := range as {
		bees[i] = f(as[i])
	}

	return bees
}

func (fs FileSystemTabCompletion) Complete(input string) (candidates []string, err error) {
	if input == "" {
		err = errors.New("no input")
		return
	}

	var info os.FileInfo
	path := input

	path, err = expandAsMuchAsPossible(path, fs.pathSep)
	if err != nil {

		if !os.IsNotExist(err) {
			return
		}

		path = filepath.Dir(path)
		_, err = os.Stat(path)
		if err != nil {
			path = ""
			return
		}
	}

	info, err = os.Stat(path)

	// this shouldn't happen unless the path is deleted after we do expansion
	if err != nil {
		return
	}

	if !info.IsDir() {
		err = errors.New("not a directory to complete")
		candidates = []string{}
		return
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return
	}

	candidates = make([]string, len(entries))

	for i, entry := range entries {
		if entry.IsDir() {
			candidates[i] = entry.Name() + fs.pathSep
		} else {
			candidates[i] = entry.Name()
		}
	}

	err = nil
	return
}

func expandAsMuchAsPossible(input, pathSep string) (path string, err error) {
	path = input

	if string(input[0]) == "~" {
		path, err = homedir.Expand(input)
		if err != nil {
			err = fmt.Errorf("bad expansion %w", err)
			return
		}
	}

	_, err = os.Stat(path)
	return
}

func (fs FileSystemTabCompletion) Rank(input string, candidates []string) []string {
	if _, err := expandAsMuchAsPossible(input, fs.pathSep); err == nil {
		return candidates
	}

	input = filepath.Base(input)
	matches := fuzzy.Find(input, candidates)
	sort.Stable(matches)

	return mapf(matches, func(m fuzzy.Match) string { return m.Str })
}

func (fs FileSystemTabCompletion) Join(current, selected string) string {
	if _, err := os.Stat(current); err != nil {
		current = filepath.Dir(current)
	}

	joined := filepath.Join(current, selected)

	// re attach directory marker if necessary
	if strings.HasSuffix(selected, fs.pathSep) {
		joined += fs.pathSep
	}

	return joined
}

func normalizePath(input, sep string) (path, trailer string) {
	if strings.HasSuffix(input, sep) {
		path = input
		return
	}

	path, trailer = filepath.Dir(input), filepath.Base(input)
	return
}
