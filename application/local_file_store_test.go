package application

import (
	"bytes"
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap/zaptest"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

const PathTest = "./"
const AppendableSuffix = ".part"
const AppendableSuffixPattern = "*" + AppendableSuffix

var (
	UsernameTest    = uuid.New().String()
	NonExistentTest = uuid.New().String()
)

func TestManager_GetFileInfo(t *testing.T) {
	// create application
	fileManager := newManager(t)

	// todo: test no ctx

	// test non-existent file size
	fileInfo, err := fileManager.GetFileInfo(newCtx(), NonExistentTest)
	if err != nil {
		t.Error("Error while running test", err)
	}
	if fileInfo.Size != 0 {
		t.Errorf("Non-zero size on non-existent file: expected %d, got %d", 0, fileInfo.Size)
	}

	// test existent appendable file
	prepareUserDir(t)
	defer cleanUserDir(t)

	file, written := prepareAppendingFile(t)

	fileInfo, err = fileManager.GetFileInfo(newCtx(), baseNoExt(file.Name()))
	if err != nil {
		t.Error("Error while running test", err)
	}
	if fileInfo.Size != int64(written) {
		t.Errorf("Bad size on existent file: expected %d, got %d", written, fileInfo.Size)
	}
}

func TestManager_AppendFile(t *testing.T) {
	fileManager := newManager(t)

	// todo: test no ctx

	// todo: handle *.part filenames

	// test append to new file, user folder not exist
	err := fileManager.AppendFile(newCtx(), NonExistentTest, newNopCloser(t, 100))
	if err != nil {
		t.Error("Error while running test", err)
	}

	// todo: test size

	// test append to appendable file
	defer cleanUserDir(t)

	file, _ := prepareAppendingFile(t)

	err = fileManager.AppendFile(newCtx(), baseNoExt(file.Name()), newNopCloser(t, 100))
	if err != nil {
		t.Error("Error while running test", err)
	}
	// todo: test size

	// test append to closed file
	file, _ = prepareClosedFile(t)

	err = fileManager.AppendFile(newCtx(), path.Base(file.Name()), newNopCloser(t, 100))
	if err != ErrConflict {
		t.Error("Error while running test: ", err)
	}
}

func TestManager_CloseFile(t *testing.T) {
	fileManager := newManager(t)

	// todo: test no ctx

	// test closing non-existent file
	err := fileManager.CloseFile(newCtx(), NonExistentTest)
	if err != nil {
		t.Error("Error while running test", err)
	}

	// test closing appending file
	prepareUserDir(t)
	defer cleanUserDir(t)

	file, _ := prepareAppendingFile(t)

	err = fileManager.CloseFile(newCtx(), baseNoExt(file.Name()))
	if err != nil {
		t.Error("Error while running test", err)
	}

	// test closing closed file
	file, _ = prepareClosedFile(t)

	err = fileManager.CloseFile(newCtx(), baseNoExt(file.Name()))
	if err != nil {
		t.Error("Error while running test", err)
	}
}

func newManager(t *testing.T) *LocalFileStore {
	fileManager, err := NewLocalFileStore(LocalFileStoreConfig{
		Path: PathTest,
	}, zaptest.NewLogger(t))
	if err != nil {
		t.Error("Error while running test", err)
	}

	return fileManager
}

func newCtx() context.Context {
	return NewContext(context.TODO(), &User{
		Username: UsernameTest,
	})
}

func prepareUserDir(t *testing.T) {
	err := os.Mkdir(filepath.Join(PathTest, UsernameTest), os.ModeDir)
	if err != nil {
		t.Error("Error while running test", err)
	}
}

func cleanUserDir(t *testing.T) {
	err := os.RemoveAll(filepath.Join(PathTest, UsernameTest))
	if err != nil {
		t.Error("Error while running test", err)
	}
}

func prepareAppendingFile(t *testing.T) (*os.File, int) {
	return newFile(t, filepath.Join(PathTest, UsernameTest), AppendableSuffixPattern, 100)
}

func prepareClosedFile(t *testing.T) (*os.File, int) {
	return newFile(t, filepath.Join(PathTest, UsernameTest), "", 100)
}

func newFile(t *testing.T, dir string, pattern string, size int64) (*os.File, int) {
	file, err := ioutil.TempFile(dir, pattern)
	if err != nil {
		t.Error("Error while running test", err)
	}

	data := make([]byte, size)

	if _, err := rand.Read(data); err != nil {
		t.Error("Error while running test", err)
	}

	n, err := file.Write(data)
	if err != nil {
		t.Error("Error while running test", err)
	}

	return file, n
}

func newNopCloser(t *testing.T, size int64) io.ReadCloser {
	data := make([]byte, size)

	_, err := rand.Read(data)
	if err != nil {
		t.Error("Error while running test", err)
	}

	return ioutil.NopCloser(bytes.NewReader(data))
}

func baseNoExt(file string) string {
	return fileNoExt(filepath.Base(file))
}

func fileNoExt(file string) string {
	return strings.TrimSuffix(file, filepath.Ext(file))
}
