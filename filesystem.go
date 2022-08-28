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

func UseFileSystemCompleter() ConfigureModel {
	return UseCompleter(NewFileSystemTabCompletion())
}

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
		//TODO(ANR): set to "." -- this causes a panic in Rank's
		//expandAsMuchAsPossible call though
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

	base := filepath.Base(input)
	matches := fuzzy.Find(base, candidates)
	sort.Stable(matches)

	candidates = mapf(matches, func(m fuzzy.Match) string { return m.Str })

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
