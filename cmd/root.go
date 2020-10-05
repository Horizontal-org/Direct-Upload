package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

// cmd args
var rpcAddress string
var verbose bool

var rootCmd = &cobra.Command{
	Use:   "direct-upload",
	Short: "Tella Direct Upload server",
	Long: `Upload server for Tella documentation app for Android. Tella is designed to protect users 
in repressive environments, it is used by activists, journalists, and civil society 
groups to document human rights violations, corruption, or electoral fraud. Tella 
encrypts and hides sensitive material on your device, and quickly deletes it in 
emergency situations; and groups and organizations can deploy it among their members 
to collect data for research, advocacy, or legal proceedings.`,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&rpcAddress, rpcFlagName, "r", "127.0.0.1:1206",
		"address for rpc server to bind to")

	rootCmd.PersistentFlags().BoolVarP(&verbose, verboseFlagName, "v", false,
		"make logging more talkative")

	//noinspection GoUnhandledErrorResult
	viper.BindPFlags(rootCmd.Flags())
}

func initConfig() {
	viper.AddConfigPath("./")
	viper.SetConfigName("config")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Can't read config:", err)
		os.Exit(1)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func isVerbose(cmd *cobra.Command) bool {
	verbose, _ := cmd.Flags().GetBool(verboseFlagName)
	return verbose || viper.GetBool(verboseFlagName)
}
