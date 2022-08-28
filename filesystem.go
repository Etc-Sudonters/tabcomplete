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

type IncludeHidden bool

const (
	IncludeHiddenFiles IncludeHidden = true
	ExcludeHiddenFiles IncludeHidden = false
)

func UseFileSystemCompleter(includeHidden IncludeHidden) ConfigureModel {
	return UseCompleter(NewFileSystemTabCompletion(bool(includeHidden)))
}

type FileSystemTabCompletion struct {
	pathSep       string
	IncludeHidden bool
}

func NewFileSystemTabCompletion(includeHidden bool) FileSystemTabCompletion {
	return FileSystemTabCompletion{
		pathSep:       string(os.PathSeparator),
		IncludeHidden: includeHidden,
	}
}

func mapf[A, B any](as []A, f func(A) B) []B {
	bees := make([]B, len(as))

	for i := range as {
		bees[i] = f(as[i])
	}

	return bees
}

func (fs FileSystemTabCompletion) Complete(input string) ([]string, error) {
	var err error
	var path string = input

	if path == "" {
		path, err = os.Getwd()
		if err != nil {
			return []string{}, err
		}
	}

	absPath, err := expandPath(path)
	if err != nil {
		return []string{}, fmt.Errorf("could not expand %s: %w", path, err)
	}

	dir, base, err := normalize(absPath)

	if err != nil {
		return []string{}, fmt.Errorf("could not normalize %s: %w", absPath, err)
	}

	entries, err := os.ReadDir(dir)

	if err != nil {
		return []string{}, fmt.Errorf("could not read %s: %w", dir, err)
	}

	candidates := mapf(entries, func(d os.DirEntry) string {
		name := d.Name()

		if d.IsDir() && !strings.HasSuffix(name, fs.pathSep) {
			return name + fs.pathSep
		}

		return name
	})

	if base != "" {
		matches := fuzzy.Find(base, candidates)
		sort.Stable(matches)
		candidates = mapf(matches, func(m fuzzy.Match) string {
			return m.Str
		})
	}

	return candidates, nil
}

func (fs FileSystemTabCompletion) Join(current, selected string) string {
	if current == "" {
		info, err := os.Stat(selected)
		if err == nil && info.IsDir() {
			return selected + fs.pathSep
		}
		return selected
	}

	expanded, err := expandPath(current)

	if err != nil {
		return filepath.Join(current, selected)
	}

	dir, _, err := normalize(expanded)

	if err == nil {
		current = dir
	}

	return filepath.Join(current, selected)
}

func normalize(path string) (string, string, error) {
	info, err := os.Stat(path)

	if err == nil {
		if !info.IsDir() {
			return filepath.Dir(path), filepath.Base(path), nil
		}

		return path, "", nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return "", "", fmt.Errorf("could not stat %s: %w", path, err)
	}

	dir, base := filepath.Split(path)

	_, err = os.Stat(dir)

	if err != nil {
		return "", "", fmt.Errorf("could not stat %s: %w", dir, err)
	}

	return dir, base, nil
}

func expandPath(input string) (path string, err error) {
	path = input

	if string(input[0]) == "~" {
		path, err = homedir.Expand(input)
		if err != nil {
			err = fmt.Errorf("bad expansion %w", err)
			return
		}
	}

	if !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		if err != nil {
			err = fmt.Errorf("bad expansion %w", err)
		}
	}

	return
}
