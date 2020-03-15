package repository

import (
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/horizontal-org/tus/application"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
)

type UserRepoConfig struct {
	DB *bolt.DB
}

type UserRepo struct {
	config UserRepoConfig

	logger *zap.Logger
	db     *bolt.DB
}

var (
	userAuthBucket = []byte("UserAuth")
	errNotFound    = errors.New("not found")
)

func NewUserRepo(config UserRepoConfig, logger *zap.Logger) (*UserRepo, error) {
	userRepo := &UserRepo{
		logger: logger,
		db:     config.DB,
	}

	err := userRepo.setupDb()
	if err != nil {
		return nil, err
	}

	return userRepo, nil
}

func (r *UserRepo) Create(user *application.UserAuth) error {
	return r.Update(user)
}

func (r *UserRepo) Read(username string) (*application.UserAuth, error) {
	var userAuth application.UserAuth

	err := r.db.View(func(tx *bolt.Tx) error {
		var v []byte

		v = tx.Bucket(userAuthBucket).Get([]byte(username))

		if v == nil {
			return errNotFound
		}

		return gob.NewDecoder(bytes.NewReader(v)).Decode(&userAuth)
	})
	if err == errNotFound {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &userAuth, nil
}

func (r *UserRepo) Update(user *application.UserAuth) error {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	err := encoder.Encode(user)
	if err != nil {
		return err
	}

	return r.db.Update(func(tx *bolt.Tx) error {
		r.logger.Debug("Update UserAuth in DB", zap.String("username", user.Username))

		return tx.Bucket(userAuthBucket).Put([]byte(user.Username), buf.Bytes())
	})
}

func (r *UserRepo) Delete(username string) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		r.logger.Debug("Delete UserAuth in DB", zap.String("username", username))

		return tx.Bucket(userAuthBucket).Delete([]byte(username))
	})
}

func (r *UserRepo) List() <-chan application.UserAuth {
	out := make(chan application.UserAuth)

	go func() {
		defer close(out)

		err := r.db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(userAuthBucket)
			c := b.Cursor()

			for k, v := c.First(); k != nil; k, v = c.Next() {
				var userAuth application.UserAuth

				err := gob.NewDecoder(bytes.NewReader(v)).Decode(&userAuth)
				if err != nil {
					return err
				}

				out <- userAuth
			}

			return nil
		})

		if err != nil {
			r.logger.Error("Error iterating bucket",
				zap.String("bucket", string(userAuthBucket)),
				zap.Error(err))
		}
	}()

	return out
}

func (r *UserRepo) setupDb() error {
	return r.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(userAuthBucket)
		return err
	})
}
