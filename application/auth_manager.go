package application

import (
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthManager struct {
	logger   *zap.Logger
	authRepo AuthRepository
}

type AuthRepository interface {
	Create(u *UserAuth) error
	Read(username string) (*UserAuth, error)
	Update(user *UserAuth) error
	Delete(username string) error
	List() <-chan UserAuth
}

func NewAuthManager(logger *zap.Logger, authRepo AuthRepository) *AuthManager {
	return &AuthManager{
		logger:   logger,
		authRepo: authRepo,
	}
}

func (m *AuthManager) HasUsername(username string) (bool, error) {
	userAuth, err := m.authRepo.Read(username)
	if err != nil {
		return false, err
	}

	return userAuth != nil, nil
}

func (m *AuthManager) SetPassword(username, password string) error {
	hash, err := m.hashPassword(password)
	if err != nil {
		return err
	}

	return m.authRepo.Create(&UserAuth{
		Username:     username,
		PasswordHash: hash,
	})
}

func (m *AuthManager) Delete(username string) error {
	return m.authRepo.Delete(username)
}

func (m *AuthManager) ListUsernames() ([]string, error) {
	var usernames []string

	for userAuth := range m.authRepo.List() {
		usernames = append(usernames, userAuth.Username)
	}

	return usernames, nil
}

func (m *AuthManager) CheckPassword(username, password string) (bool, error) {
	userAuth, err := m.authRepo.Read(username)
	if userAuth == nil || err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(userAuth.PasswordHash), []byte(password))
	if err != nil {
		return false, err
	}

	return true, nil
}

func (m *AuthManager) hashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}
