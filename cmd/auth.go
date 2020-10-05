package cmd

import (
	"errors"
	"fmt"
	"github.com/horizontal-org/direct-upload/application"
	logger2 "github.com/horizontal-org/direct-upload/logger"
	rpcSrv "github.com/horizontal-org/direct-upload/server/rpc"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh/terminal"
	"net/rpc"
	"syscall"
)

type consumer func(logger *zap.Logger, client *rpc.Client) error

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage user authentication.",
}

var authAddCmd = &cobra.Command{
	Use:   "add <username>",
	Short: "Add user authentication if doesn't already exists. Will prompt for password.",
	Args:  cobra.ExactArgs(1),
	RunE:  authAddCmdFunc,
}

var authDelCmd = &cobra.Command{
	Use:   "del <username>",
	Short: "Delete user authentication.",
	Args:  cobra.ExactArgs(1),
	RunE:  authDelCmdFunc,
}

var authChangePassCmd = &cobra.Command{
	Use:   "passwd <username>",
	Short: "Change user authentication. Will prompt for password.",
	Args:  cobra.ExactArgs(1),
	RunE:  authPasswdCmdFunc,
}

var authListCmd = &cobra.Command{
	Use:   "list",
	Short: "List usernames.",
	Args:  cobra.ExactArgs(0),
	RunE:  authListCmdFunc,
}

var authBackupCmd = &cobra.Command{
	Use:   "backup <path>",
	Short: "Backup auth database.",
	Args:  cobra.ExactArgs(1),
	RunE:  authBackupCmdFunc,
}

var errUsernameNotValid = errors.New("username not valid")
var errUsernameExists = errors.New("username exists")

//noinspection GoUnhandledErrorResult
func init() {
	authCmd.AddCommand(authAddCmd)
	authCmd.AddCommand(authDelCmd)
	authCmd.AddCommand(authChangePassCmd)
	authCmd.AddCommand(authListCmd)
	authCmd.AddCommand(authBackupCmd)
	rootCmd.AddCommand(authCmd)
}

//noinspection GoUnusedParameter
func authAddCmdFunc(cmd *cobra.Command, args []string) error {
	return with(cmd, func(logger *zap.Logger, client *rpc.Client) error {
		username := args[0]

		if !application.ValidUsername(username) {
			return errUsernameNotValid
		}

		hasRequest := &rpcSrv.HasUsernameRequest{
			Username: username,
		}

		var exists bool

		logger.Debug("Calling RpcServer.AddAuth", zap.String("username", hasRequest.Username))

		err := client.Call("RpcServer.HasUsername", hasRequest, &exists)
		if err != nil {
			return err
		}

		if exists {
			return errUsernameExists
		}

		password, err := readPassword()
		if err != nil {
			return err
		}

		if password == "" {
			return nil
		}

		addRequest := &rpcSrv.AddAuthRequest{
			Username: username,
			Password: password,
		}

		var reply rpcSrv.Response

		logger.Debug("Calling RpcServer.AddAuth", zap.String("username", addRequest.Username))

		err = client.Call("RpcServer.AddAuth", addRequest, &reply)
		if err == rpcSrv.ErrUsernameExists {
			fmt.Println(err)
			return nil
		}

		return err
	})
}

//noinspection GoUnusedParameter
func authDelCmdFunc(cmd *cobra.Command, args []string) error {
	return with(cmd, func(logger *zap.Logger, client *rpc.Client) error {
		username := args[0]

		if !application.ValidUsername(username) {
			return errUsernameNotValid
		}

		delRequest := &rpcSrv.DelAuthRequest{
			Username: username,
		}

		var reply rpcSrv.Response

		logger.Debug("Calling RpcServer.DelAuth", zap.String("username", delRequest.Username))

		return client.Call("RpcServer.DelAuth", delRequest, &reply)
	})
}

//noinspection GoUnusedParameter
func authPasswdCmdFunc(cmd *cobra.Command, args []string) error {
	return with(cmd, func(logger *zap.Logger, client *rpc.Client) error {
		username := args[0]

		if !application.ValidUsername(username) {
			return errUsernameNotValid
		}

		password, err := readPassword()
		if err != nil {
			return err
		}

		if password == "" {
			return nil
		}

		setRequest := &rpcSrv.SetAuthRequest{
			Username: username,
			Password: password,
		}

		var reply rpcSrv.Response

		logger.Debug("Calling RpcServer.SetAuth", zap.String("username", setRequest.Username))

		return client.Call("RpcServer.SetAuth", setRequest, &reply)
	})
}

//noinspection GoUnusedParameter
func authListCmdFunc(cmd *cobra.Command, args []string) error {
	return with(cmd, func(logger *zap.Logger, client *rpc.Client) error {
		var reply []string

		logger.Debug("Calling RpcServer.ListUsernames")

		err := client.Call("RpcServer.ListUsernames", &rpcSrv.Request{}, &reply)
		if err != nil {
			return err
		}

		for _, username := range reply {
			fmt.Println(username)
		}

		return nil
	})
}

//noinspection GoUnusedParameter
func authBackupCmdFunc(cmd *cobra.Command, args []string) error {
	return with(cmd, func(logger *zap.Logger, client *rpc.Client) error {
		path := args[0]

		backupRequest := &rpcSrv.BackupAuthRequest{
			Path: path,
		}

		var reply rpcSrv.Response

		logger.Debug("Calling RpcServer.BackupDatabase")

		return client.Call("RpcServer.BackupDatabase", &backupRequest, &reply)
	})
}

func with(cmd *cobra.Command, fn consumer) error {
	logger, _ := logger2.NewLogger(isVerbose(cmd))
	//goland:noinspection GoUnhandledErrorResult
	defer logger.Sync()

	rpcAddress, err := cmd.Flags().GetString(rpcFlagName)
	if err != nil {
		return err
	}

	logger.Debug("Connecting to rpc server", zap.String("address", rpcAddress))

	rcpClient, err := rpc.Dial("tcp", rpcAddress)
	if err != nil {
		logger.Fatal("Unable to create rpc client", zap.Error(err))
	}

	return fn(logger, rcpClient)
}

func readPassword() (string, error) {
	fmt.Print("Password: ")

	bytes, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	fmt.Println()

	return string(bytes), nil
}
