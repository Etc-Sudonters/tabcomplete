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

var (
	ErrCannotGetWorkingDirectory = errors.New("cannot get working directory")
	ErrCouldNotExpandPath        = errors.New("could not expand path")
	ErrCouldNotExpandHome        = errors.New("could not expand home directory")
	ErrCouldNotNormalizePath     = errors.New("could not normalize path")
	ErrCouldNotReadDir           = errors.New("could not read directory")
)

type FileSystemError struct {
	reason error
	err    error
}

func (f FileSystemError) Is(target error) bool {
	return errors.Is(f.reason, target)
}

func (f FileSystemError) Error() string {
	return fmt.Sprintf("%s: %s", f.reason, f.err.Error())
}

func (f FileSystemError) Unwrap() error {
	return f.err
}

func UseFileSystemCompleter() ConfigureModel {
	return UseCompleter(NewFileSystemTabCompletion())
}

type FileSystemTabCompletion struct {
	pathSep string
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

func (fs FileSystemTabCompletion) Complete(input string) ([]string, error) {
	var err error
	var path string = input

	if path == "" {
		path, err = os.Getwd()
		if err != nil {
			return []string{}, FileSystemError{
				reason: ErrCannotGetWorkingDirectory,
				err:    err,
			}
		}
	}

	absPath, err := expandPath(path)
	if err != nil {
		return []string{}, err
	}

	dir, base := normalize(absPath)

	entries, err := os.ReadDir(dir)

	if err != nil {
		return []string{}, FileSystemError{
			reason: fmt.Errorf("%w: %s", ErrCouldNotReadDir, dir),
			err:    err,
		}
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
	current, _ = normalize(current)

	return filepath.Join(current, selected)
}

func normalize(path string) (dir string, base string) {
	info, err := os.Stat(path)

	if err == nil {
		if !info.IsDir() {
			dir, base := filepath.Split(path)
			return dir, base
		}

		return path, ""
	}

	dir, base = filepath.Split(path)

	if dir == "" && (base == "." || base == "~") {
		return base, dir
	}

	return dir, base
}

func expandPath(input string) (path string, err error) {
	path = input

	if string(input[0]) == "~" {
		path, err = homedir.Expand(input)
		if err != nil {
			err = FileSystemError{
				reason: ErrCouldNotExpandHome,
				err:    err,
			}
			return
		}
	}

	if !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		if err != nil {
			err = FileSystemError{
				reason: fmt.Errorf("%w: %s", ErrCouldNotExpandPath, path),
				err:    err,
			}
		}
	}

	return
}
