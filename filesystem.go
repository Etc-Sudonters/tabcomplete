package tabcomplete

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/sahilm/fuzzy"
)

type fileSystemTabCompletion struct {
	pathSep string
}

func NewFileSystemTabCompletion() TabCompletion {
	return fileSystemTabCompletion{
		pathSep: string(os.PathListSeparator),
	}
}

func (fs fileSystemTabCompletion) Complete(input string) (candidates []string) {
	expanded, err := homedir.Expand(input)

	if err != nil {
		return
	}

	path, _ := normalizePath(expanded, fs.pathSep)
	entries, err := os.ReadDir(path)

	if err != nil {
		return
	}

	candidates = make([]string, 0, len(entries))

	for _, entry := range entries {
		name := entry.Name()

		if entry.IsDir() {
			candidates = append(candidates, name+fs.pathSep)
		} else {
			candidates = append(candidates, name)
		}

	}
	return
}

func (fs fileSystemTabCompletion) Rank(input string, candidates []string) (rankedCandidates []string) {
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

func (fs fileSystemTabCompletion) Join(current, selected string) string {
	return filepath.Join(current, selected)
}

func normalizePath(input, sep string) (path, trailer string) {
	if strings.HasSuffix(input, sep) {
		path = input
		return
	}

	path, trailer = filepath.Dir(input), filepath.Base(input)
	return
}
