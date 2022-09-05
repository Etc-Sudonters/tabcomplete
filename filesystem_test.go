package tabcomplete

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSystemComplete(t *testing.T) {
	tempDir, err := populateTempDir(t, []string{"README.md", "examples.go", "kenosis.lrc", "bookclub", ".todo", ".hidden"})
	require.Nil(t, err)
	fsCompleter := NewFileSystemTabCompletion()

	candidates, err := fsCompleter.Complete(filepath.Join(tempDir, "e"))

	require.Nil(t, err)
	require.ElementsMatch(t, []string{"README.md", "examples.go", "kenosis.lrc", ".hidden"}, candidates)
	assert.Equal(t, "examples.go", candidates[0])
	assert.Equal(t, "README.md", candidates[1])
	assert.Equal(t, "kenosis.lrc", candidates[2])
	assert.Equal(t, ".hidden", candidates[3])
}

func TestFileSystemCompleter_UsesCurrentDirectory_ForUnqualifiedPaths(t *testing.T) {
	tempDir, err := populateTempDir(t, []string{"README.md", "examples.go", "kenosis.lrc", "bookclub", ".todo", ".hidden"})
	require.Nil(t, err)
	require.Nil(t, os.Chdir(tempDir))

	fsCompleter := NewFileSystemTabCompletion()
	candidates, err := fsCompleter.Complete("e")

	require.Nil(t, err)
	assert.ElementsMatch(t, []string{"README.md", "examples.go", "kenosis.lrc", ".hidden"}, candidates)
	assert.Equal(t, "examples.go", candidates[0])
	assert.Equal(t, "README.md", candidates[1])
	assert.Equal(t, "kenosis.lrc", candidates[2])
	assert.Equal(t, ".hidden", candidates[3])
}

func TestFileSystemCompleter_ListsAllFiles_IfJustGivenDir(t *testing.T) {
	fileNames := []string{"README.md", "examples.go", "kenosis.lrc", "bookclub", ".todo", ".hidden"}
	tempDir, err := populateTempDir(t, fileNames)
	require.Nil(t, err)

	fsCompleter := NewFileSystemTabCompletion()
	candidates, err := fsCompleter.Complete(tempDir)

	require.Nil(t, err)
	require.ElementsMatch(t, fileNames, candidates)
}

func TestFileSystemCompleter_ListsAllFiles_InCurrentDirectory_IfEmptyPath(t *testing.T) {
	fileNames := []string{"V", "Nebula", "Hibernate", "Pamela", "All Class"}
	tempDir, err := populateTempDir(t, fileNames)
	require.Nil(t, err)
	require.Nil(t, os.Chdir(tempDir))

	fsCompleter := NewFileSystemTabCompletion()

	candidates, err := fsCompleter.Complete("")

	require.Nil(t, err)
	require.ElementsMatch(t, fileNames, candidates)
}

func TestFileSystemCompleter_ExpandsTilde_ToHomeDir(t *testing.T) {
	tempDir, err := populateTempDir(t, []string{"README.md", "examples.go", "kenosis.lrc", "bookclub", ".todo", ".hidden"})
	require.Nil(t, err)

	os.Setenv("HOME", tempDir)

	fsCompleter := NewFileSystemTabCompletion()
	candidates, err := fsCompleter.Complete("~/e")

	require.Nil(t, err)
	require.ElementsMatch(t, []string{"README.md", "examples.go", "kenosis.lrc", ".hidden"}, candidates)

	assert.Equal(t, "examples.go", candidates[0])
	assert.Equal(t, "README.md", candidates[1])
	assert.Equal(t, "kenosis.lrc", candidates[2])
	assert.Equal(t, ".hidden", candidates[3])
}

func TestFileSystemCompleter_AppendsPathSep_ToDirectoryNames(t *testing.T) {
	fileNames := []string{"README.md", "examples" + string(os.PathSeparator), "Nebula", "Hibernate", "Pamela" + string(os.PathSeparator)}
	tempDir, err := populateTempDir(t, fileNames)
	require.Nil(t, err)

	fsCompleter := NewFileSystemTabCompletion()

	candidates, err := fsCompleter.Complete(tempDir)

	require.Nil(t, err)
	require.ElementsMatch(t, fileNames, candidates)
}

func TestFileSystemCompleter_ReturnsOnlyFilename_IfProvidedPathExists(t *testing.T) {
	tempDir, err := populateTempDir(t, []string{"README.md", "examples.go", "kenosis.lrc", "bookclub", ".todo", ".hidden"})
	require.Nil(t, err)

	fsCompleter := NewFileSystemTabCompletion()
	candidates, err := fsCompleter.Complete(filepath.Join(tempDir, "README.md"))

	require.Nil(t, err)
	require.Equal(t, 1, len(candidates))
	require.Equal(t, "README.md", candidates[0])
}

func TestFileSystemCompleter_Fails_IfPathIsEmpty_AndCantGetCurrentDirectory(t *testing.T) {
	tempDir := t.TempDir()
	require.Nil(t, os.Chdir(tempDir))
	require.Nil(t, os.Remove(tempDir))

	fsCompleter := NewFileSystemTabCompletion()
	candidates, err := fsCompleter.Complete("")
	require.Empty(t, candidates)
	require.NotNil(t, err)

	var fsError FileSystemError
	require.ErrorAs(t, err, &fsError)
	assert.ErrorIs(t, fsError, ErrCannotGetWorkingDirectory)
}

func TestFileSystemCompleter_FailsIfCannotExpandPath_BecauseDirDoesNotExist(t *testing.T) {
	madeUpDir := uuid.New()

	fsCompleter := NewFileSystemTabCompletion()

	candidates, err := fsCompleter.Complete(filepath.Join(madeUpDir.String(), "anywhere"))

	require.Empty(t, candidates)
	require.NotNil(t, err)

	var fsError FileSystemError
	require.ErrorAs(t, err, &fsError)
	assert.ErrorIs(t, fsError, ErrCouldNotExpandPath)
}

func TestFilesystemCompleter_Complete_FollowsUpDirectoryRepresentation(t *testing.T) {
	tempDir, err := populateTempDir(t, []string{"Immolation" + string(os.PathSeparator), "Thornhill" + string(os.PathSeparator)})
	require.Nil(t, err)
	err = os.Chdir(filepath.Join(tempDir, "Immolation"))
	require.Nil(t, err, "%s", err)

	fsCompleter := NewFileSystemTabCompletion()

	candidates, err := fsCompleter.Complete("..")
	require.Nil(t, err)
	require.ElementsMatch(t, []string{"Immolation" + string(os.PathSeparator), "Thornhill" + string(os.PathSeparator)}, candidates)
}

func TestFileSystemCompleter_Complete_FollowsCurrentDirectoryRepresentation(t *testing.T) {
	tempDir, err := populateTempDir(t, []string{"downfall", "pareidolia"})
	require.Nil(t, err)
	require.Nil(t, os.Chdir(tempDir))

	fsCompleter := NewFileSystemTabCompletion()

	candidates, err := fsCompleter.Complete("." + string(os.PathSeparator) + "a")

	require.Nil(t, err)
	require.ElementsMatch(t, []string{"downfall", "pareidolia"}, candidates)
}

func populateTempDir(t testing.TB, names []string) (string, error) {
	tempDir := t.TempDir()

	var err error
	for _, fn := range names {

		joined := filepath.Join(tempDir, fn)

		if shouldBeDir(fn) {
			err = os.Mkdir(joined, 0700)
		} else {
			_, err = os.Create(joined)
		}

		if err != nil {
			return tempDir, err
		}
	}

	return tempDir, nil
}

func TestFileSystemCompleter_FailsIfCouldNotReadTargetDirectory(t *testing.T) {
	tempDir := t.TempDir()
	madeUpDir := filepath.Join(tempDir, uuid.NewString()) + string(os.PathSeparator)

	fsCompleter := NewFileSystemTabCompletion()
	candidates, err := fsCompleter.Complete(filepath.Join(madeUpDir, "preach"))

	require.Empty(t, candidates)
	require.NotNil(t, err)

	var fsError FileSystemError
	require.ErrorAs(t, err, &fsError)
	require.ErrorIs(t, err, ErrCouldNotReadDir)
}

func TestFileSystemCompleter_Join_ReturnsSelected_IfCurrentIsEmpty(t *testing.T) {
	tempDir, err := populateTempDir(t, []string{"Ulcerborne"})
	require.Nil(t, err)
	require.Nil(t, os.Chdir(tempDir))

	fsCompleter := NewFileSystemTabCompletion()

	joined := fsCompleter.Join("", "Ulcerborne")

	require.Equal(t, "Ulcerborne", joined)
}

func TestFileSystemCompleter_Join_OverwritesLastPathSegementOnCurrent_WhenLastPathSegmentIsNotADir(t *testing.T) {
	tempDir, err := populateTempDir(t, []string{"Ulcerborne"})
	require.Nil(t, err)

	fsCompleter := NewFileSystemTabCompletion()

	joined := fsCompleter.Join(filepath.Join(tempDir, "ulc"), "Ulcerborne")

	require.Equal(t, filepath.Join(tempDir, "Ulcerborne"), joined)
}

func TestFileSystemCompleter_Join_PreservesCurrentDirectoryRepresentation(t *testing.T) {
	tempDir, err := populateTempDir(t, []string{"EntrancedByCalamity"})
	require.Nil(t, err)
	require.Nil(t, os.Chdir(tempDir))

	fsCompleter := NewFileSystemTabCompletion()

	joined := fsCompleter.Join(
		filepath.Join(".", "Entra"),
		"EntrancedByCalamity",
	)

	require.Equal(t, filepath.Join(".", "EntrancedByCalamity"), joined)
}

func TestFileSystemCompletere_Join_PreservesHomeDirectoryRepresentation(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)

	fsCompleter := NewFileSystemTabCompletion()

	joined := fsCompleter.Join("~/Ulc", "Ulcerborne")
	require.Equal(t, "~/Ulcerborne", joined)
}

func TestFileSystemCompleter_Join_PreservesUpDirectoryRepresentation(t *testing.T) {
	fsCompleter := NewFileSystemTabCompletion()

	joined := fsCompleter.Join(".."+string(os.PathSeparator)+"T", "Thornhill")
	require.Equal(t, ".."+string(os.PathSeparator)+"Thornhill", joined)
}

func TestFileSystemCompleter_Join_PreservesPrefixIfItIsAloneInCurrent(t *testing.T) {
	prefixes := []string{"..", "~"}

	for _, p := range prefixes {
		prefix := p

		t.Run(prefix, func(t *testing.T) {
			fsCompleter := NewFileSystemTabCompletion()
			joined := fsCompleter.Join(prefix, "GorgonSisters")
			require.Equal(t, prefix+string(os.PathSeparator)+"GorgonSisters", joined)
		})
	}
}

func shouldBeDir(fp string) bool {
	return strings.HasSuffix(fp, string(os.PathSeparator))
}
