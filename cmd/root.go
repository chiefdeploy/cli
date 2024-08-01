package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "chief",
	Short: "Chief CLI is a command line tool for managing the Chief controller.",
	Long:  "Chief CLI is a command line tool for managing the Chief controller.",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	err := os.MkdirAll("/var/chief", 0755)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err)))
		os.Exit(1)
	}

	viper.AddConfigPath("/var/chief")
	viper.SetConfigType("yaml")
	viper.SetConfigName("chief")

	viper.SetConfigFile("/var/chief/chief.yaml")

	viper.SetDefault("domain", "")
	viper.SetDefault("automatic_updates", true)

	viper.SetDefault("current_version", "")

	viper.SafeWriteConfig()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("No config file found.")
	}
}
