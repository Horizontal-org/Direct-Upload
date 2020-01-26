package db

import (
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
	"sync"
	"time"
)

type BoltConnection struct {
	db *bolt.DB
}

var once sync.Once
var bc BoltConnection

func NewBoltConnection(logger *zap.Logger, path string) *BoltConnection {
	once.Do(func() {
		var err error

		logger.Sugar().Infof("Opening BoltDB database in %s", path)

		bc.db, err = bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
		if err != nil {
			logger.Fatal("Unable to open BoltDB", zap.String("path", path), zap.Error(err))
		}
	})

	return &bc
}

func (c *BoltConnection) GetDB() *bolt.DB {
	return c.db
}

func (c *BoltConnection) Close(logger *zap.Logger) {
	err := c.db.Close()
	if err != nil {
		logger.Error("Error closing BoltDB", zap.String("path", c.db.Path()), zap.Error(err))
	}
}

func (c *BoltConnection) Backup(logger *zap.Logger, toPath string) error {
	return c.db.View(func(tx *bolt.Tx) error {
		return tx.CopyFile(toPath, 0600)
	})
}
