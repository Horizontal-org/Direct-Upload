package application

import (
	"context"
	"go.uber.org/zap"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalFileStore struct {
	config LocalFileStoreConfig
	logger *zap.Logger
}

type LocalFileStoreConfig struct {
	Path string
}

type localFile struct {
	path   string
	size   int64
	exists bool
	closed bool
}

// todo: make this unique for every file so we can accept ".part" extensions
const appendableSuffix = ".part"

func NewLocalFileStore(config LocalFileStoreConfig, logger *zap.Logger) (*LocalFileStore, error) {
	return &LocalFileStore{
		config: config,
		logger: logger,
	}, nil
}

func (m *LocalFileStore) GetFileInfo(ctx context.Context, file string) (*FileInfo, error) {
	user, ok := UserFromContext(ctx)
	if !ok {
		return nil, ErrNoUserCtx
	}

	localFile, err := m.getLocalFile(user.Username, file)
	if err != nil {
		m.logger.Error("Error getting local file",
			zap.String("username", user.Username), zap.String("file", file), zap.Error(err))
		return nil, err
	}

	m.logger.Info("Returning file info", zap.String("file", file), zap.Int64("size", localFile.size))

	return &FileInfo{
		Size: localFile.size,
	}, nil
}

func (m *LocalFileStore) AppendFile(ctx context.Context, file string, data io.ReadCloser) error {
	user, ok := UserFromContext(ctx)
	if !ok {
		m.logger.Debug("AppendFile: no User in ctx",
			zap.String("username", user.Username), zap.String("file", file))
		return ErrNoUserCtx
	}

	m.logger.Debug("AppendFile: started",
		zap.String("username", user.Username), zap.String("file", file))

	localFile, err := m.getLocalFile(user.Username, file)
	if err != nil {
		m.logger.Error("Error getting local file",
			zap.String("username", user.Username), zap.String("file", file), zap.Error(err))
		return err
	}

	if localFile.closed {
		m.logger.Error("Appending on closed file", zap.String("file", file))
		return ErrConflict
	}

	// create dir, ignore error
	m.createUserDir(user.Username)

	out, err := os.OpenFile(localFile.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		m.logger.Error("Error opening file", zap.Error(err), zap.String("file", file))
		return err
	}
	//noinspection GoUnhandledErrorResult
	defer out.Close()

	written, err := io.Copy(out, data)
	if err != nil {
		m.logger.Error("Error writing to file", zap.Error(err), zap.String("file", file))
		return err
	}

	err = out.Sync()
	if err != nil {
		m.logger.Error("Error syncing file", zap.Error(err), zap.String("file", file))
		return err
	}

	m.logger.Info("Appending file", zap.String("file", file), zap.Int64("written", written))

	return nil
}

func (m *LocalFileStore) CloseFile(ctx context.Context, file string) error {
	user, ok := UserFromContext(ctx)
	if !ok {
		return ErrNoUserCtx
	}

	localFile, err := m.getLocalFile(user.Username, file)
	if err != nil {
		m.logger.Error("Error getting local file",
			zap.String("username", user.Username), zap.String("file", file), zap.Error(err))
		return err
	}

	if !localFile.exists {
		m.logger.Warn("Closing non-existent file", zap.String("file", file))
		return nil
	}

	if localFile.closed {
		m.logger.Info("Closing already closed file", zap.String("file", file))
		return nil
	}

	err = os.Rename(localFile.path, strings.TrimSuffix(localFile.path, appendableSuffix))
	if err != nil {
		m.logger.Error("Error renaming file", zap.Error(err), zap.String("file", file))
		return err
	}

	m.logger.Info("Closing file", zap.String("file", file))

	return nil
}

func (m *LocalFileStore) getPartName(file string) string {
	if !strings.HasSuffix(file, appendableSuffix) {
		return file + appendableSuffix
	}

	return file
}

func (m *LocalFileStore) getFullPath(username string, file string) string {
	return filepath.Join(m.getFullDir(username), file)
}

func (m *LocalFileStore) getFullDir(username string) string {
	return filepath.Join(m.config.Path, username)
}

func (m *LocalFileStore) getLocalFile(username string, file string) (*localFile, error) {
	// check if there is a closed file
	closed, err := newLocalFile(m.getFullPath(username, file), true)
	if err != nil {
		return nil, err
	}

	if closed.exists {
		return closed, nil
	}

	// closed does not exist, return current or future .part file
	part, err := newLocalFile(m.getFullPath(username, m.getPartName(file)), false)
	if err != nil {
		return nil, err
	}

	return part, nil
}

func (m *LocalFileStore) createUserDir(username string) {
	dir := m.getFullDir(username)
	_ = os.Mkdir(dir, os.ModeDir)
	m.logger.Debug("Dir created", zap.String("path", dir))
}

func newLocalFile(path string, closed bool) (*localFile, error) {
	file := &localFile{
		path:   path,
		closed: closed,
	}

	stat, err := os.Stat(file.path)

	if os.IsNotExist(err) {
		return file, nil
	}

	if err != nil {
		return nil, err
	}

	file.exists = true
	file.size = stat.Size()

	return file, nil
}
