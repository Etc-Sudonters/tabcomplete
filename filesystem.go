package tabcomplete

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
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

func (fs FileSystemTabCompletion) Complete(input string) (candidates []string, err error) {
	if len(input) == 0 {
		return
	}

	var path string
	var pathExists bool
	if input[0] == byte('~') {

		path, err = homedir.Expand(input)

		if err != nil {
			return
		}
	}

	if !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		if err != nil {
			return
		}
	}

	pathExists, err = checkIfExists(path)

	if err != nil {
		// TODO(ANR): should figure out what else could pop out
		// this is fine for now
		return
	}

	if !pathExists {
		path, _ = normalizePath(path, fs.pathSep)
		pathExists, err = checkIfExists(path)
		if err != nil {
			return
		}
		if !pathExists {
			err = os.ErrNotExist
			return
		}
	}

	entries, err := os.ReadDir(path)

	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "not a directory") {
			candidates = []string{}
			err = nil
		}
		return
	}

	candidates = make([]string, 0, len(entries))

	for _, entry := range entries {
		name := entry.Name()

		if name == "" {
			continue
		}

		if !fs.IncludeHidden && name[0] == byte('.') {
			continue
		}

		if entry.IsDir() {
			candidates = append(candidates, name+fs.pathSep)
		} else {
			candidates = append(candidates, name)
		}

	}
	return
}

func (fs FileSystemTabCompletion) Rank(input string, candidates []string) (rankedCandidates []string) {
	_, trailer := normalizePath(input, fs.pathSep)
	if trailer == "" {
		rankedCandidates = candidates
		return
	}

	matches := fuzzy.Find(trailer, candidates)
	sort.Stable(matches)

	rankedCandidates = make([]string, len(candidates))

	for i, match := range matches {
		rankedCandidates[i] = match.Str
	}

	return
}

func (fs FileSystemTabCompletion) Join(current, selected string) string {
	path := filepath.Dir(current)
	expanded, err := filepath.Abs(path)
	if err != nil {
		expanded = path
	}
	return filepath.Join(expanded, selected)
}

func normalizePath(input, sep string) (path, trailer string) {
	if strings.HasSuffix(input, sep) {
		path = input
		return
	}

	path, trailer = filepath.Dir(input), filepath.Base(input)
	return
}

func checkIfExists(path string) (exists bool, err error) {
	_, err = os.Stat(path)
	if err == nil {
		exists = true
	} else if os.IsNotExist(err) {
		exists = false
		err = nil
	} else {
		exists = false
	}
	return
}
