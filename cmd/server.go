package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tomislavr/tus/application"
	"github.com/tomislavr/tus/db"
	logger2 "github.com/tomislavr/tus/logger"
	"github.com/tomislavr/tus/repository"
	"github.com/tomislavr/tus/server/http"
	"github.com/tomislavr/tus/server/rpc"
	"go.uber.org/zap"
)

const (
	addressFlagName  = "address"
	databaseFlagName = "database"
	filesFlagName    = "files"
	certFlagName     = "cert"
	keyFlagName      = "key"
	rpcFlagName      = "rpc"
	verboseFlagName  = "verbose"
)

// cmd args
var address, database, files, cert, key string

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start Tella Upload Server",
	Run:   serverCmdFunc,
}

//noinspection GoUnhandledErrorResult
func init() {
	serverCmd.Flags().StringVarP(&address, addressFlagName, "a", ":8080",
		"address for server to bind to")

	serverCmd.Flags().StringVarP(&cert, certFlagName, "c", viper.GetString(certFlagName),
		"certificate file, ie. ./fullcert.pem")

	serverCmd.Flags().StringVarP(&database, databaseFlagName, "d", "./tus.db",
		"tus database file")

	serverCmd.Flags().StringVarP(&files, filesFlagName, "f", viper.GetString(filesFlagName),
		"path where tus stores uploaded files")

	serverCmd.Flags().StringVarP(&key, keyFlagName, "k", viper.GetString(keyFlagName),
		"private key file, ie. ./key.pem")

	viper.BindPFlags(serverCmd.Flags())

	rootCmd.AddCommand(serverCmd)
}

//noinspection GoUnusedParameter
func serverCmdFunc(cmd *cobra.Command, args []string) {
	logger, _ := logger2.NewLogger(isVerbose(cmd))
	defer logger.Sync()

	localFileStore, _ := application.NewLocalFileStore(application.LocalFileStoreConfig{
		Path: viper.GetString(filesFlagName),
	}, logger)

	conn := db.NewBoltConnection(logger, viper.GetString(databaseFlagName))
	defer conn.Close(logger)

	authRepository, err := repository.NewUserRepo(repository.UserRepoConfig{
		DB: conn.GetDB(),
	}, logger)
	if err != nil {
		logger.Fatal("Unable to create User Repository", zap.Error(err))
	}

	authManager := application.NewAuthManager(logger, authRepository)

	// start http server
	go http.NewServer(http.Config{
		Address:        viper.GetString(addressFlagName),
		CertFile:       viper.GetString(certFlagName),
		PrivateKeyFile: viper.GetString(keyFlagName),
	}, authManager, localFileStore, logger).Start()

	rpcAddress, err := cmd.Flags().GetString(rpcFlagName)
	if err != nil {
		logger.Fatal("Unknown rpc server address", zap.Error(err))
	}

	// start rpc server
	rpc.StartRpcServer(rpc.Config{
		Path: rpcAddress,
	}, authManager, conn, logger)
}
