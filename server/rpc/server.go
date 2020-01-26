package rpc

import (
	"errors"
	"github.com/tomislavr/tus/application"
	"github.com/tomislavr/tus/db"
	"go.uber.org/zap"
	"net"
	"net/rpc"
)

type Config struct {
	Path string
}

//noinspection GoNameStartsWithPackageName
type RpcServer struct {
	config Config
	am     *application.AuthManager
	bc     *db.BoltConnection
	logger *zap.Logger
}

type Response bool
type Request struct{}

type UsernameRequest struct {
	Username string
}

type AddAuthRequest struct {
	Username string
	Password string
}

type BackupAuthRequest struct {
	Path string
}

type SetAuthRequest AddAuthRequest
type DelAuthRequest UsernameRequest
type HasUsernameRequest UsernameRequest

var ErrUsernameNotValid = errors.New("username not valid")
var ErrUsernameExists = errors.New("username already exists")

func StartRpcServer(config Config, authManager *application.AuthManager, bc *db.BoltConnection, logger *zap.Logger) {
	srv := &RpcServer{
		config: config,
		am:     authManager,
		bc:     bc,
		logger: logger,
	}

	err := rpc.Register(srv)
	if err != nil {
		srv.logger.Fatal("unable to register rpc server", zap.Error(err))
	}

	listener, err := net.Listen("tcp", srv.config.Path)
	if err != nil {
		srv.logger.Fatal("rpc server unable to listen", zap.String("Path", srv.config.Path), zap.Error(err))
	}

	srv.logger.Sugar().Infof("Starting Tella RPC server on %s", config.Path)

	rpc.Accept(listener)
}

func (a *RpcServer) AddAuth(req *AddAuthRequest, res *Response) error {
	if !application.ValidUsername(req.Username) {
		return ErrUsernameNotValid
	}

	exists, err := a.am.HasUsername(req.Username)
	if err != nil {
		return err
	}

	if exists {
		return ErrUsernameExists
	}

	return a.am.SetPassword(req.Username, req.Password)
}

func (a *RpcServer) DelAuth(req *DelAuthRequest, res *Response) error {
	if !application.ValidUsername(req.Username) {
		return ErrUsernameNotValid
	}

	return a.am.Delete(req.Username)
}

func (a *RpcServer) SetAuth(req *SetAuthRequest, res *Response) error {
	if !application.ValidUsername(req.Username) {
		return ErrUsernameNotValid
	}

	return a.am.SetPassword(req.Username, req.Password)
}

func (a *RpcServer) ListUsernames(req *Request, res *[]string) error {
	usernames, err := a.am.ListUsernames()
	if err != nil {
		return err
	}

	*res = usernames

	return nil
}

func (a *RpcServer) HasUsername(req *HasUsernameRequest, res *bool) error {
	ok, err := a.am.HasUsername(req.Username)
	if err != nil {
		return err
	}

	*res = ok

	return nil
}

func (a *RpcServer) BackupDatabase(req *BackupAuthRequest, res *Response) error {
	return a.bc.Backup(a.logger, req.Path)
}
