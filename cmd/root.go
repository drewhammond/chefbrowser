package cmd

import (
	"fmt"
	"github.com/drewhammond/chefbrowser/internal/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "chefbrowser",
	Short: "A web application for viewing chef server resources",
	Run: func(cmd *cobra.Command, args []string) {
		app.New()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.chefbrowser.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	if cfgFile != "" {
		viper.SetConfigType("toml")
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("chefbrowser")
		viper.SetConfigType("toml")
		viper.AddConfigPath(".")
	}

	viper.AutomaticEnv() // read in environment variables that match

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// config file not found
			fmt.Println("config file not found")
			os.Exit(1)
		} else {
			// found, but another error...
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
