package cmd

import (
	"bytes"
	"fmt"
	"os"

	"github.com/drewhammond/chefbrowser/config"
	"github.com/drewhammond/chefbrowser/internal/app"
	"github.com/drewhammond/chefbrowser/internal/common/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/ini.v1"
)

var (
	cfgFile string
	cfg     config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "chefbrowser",
	Short:   "A web application for viewing chef server resources",
	Version: version.Get().Version, // todo: format this
	Run: func(cmd *cobra.Command, args []string) {
		app.New(&cfg)
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "path to config file")
}

// initConfig reads in config defaults, user config files, and ENV variables if set.
func initConfig() {
	// ignore inline comments to allow # characters in the middle of custom links and other properties - (#399)
	v := viper.NewWithOptions(viper.IniLoadOptions(ini.LoadOptions{IgnoreInlineComment: true}))
	v.SetConfigType("ini")
	v.SetConfigName("chefbrowser")
	v.AddConfigPath("/etc/chefbrowser/")

	// load defaults
	err := v.ReadConfig(bytes.NewBuffer(config.DefaultConfig))
	if err != nil {
		fmt.Println("failed to read default config, err:", err)
		os.Exit(1)
	}

	if cfgFile != "" {
		v.SetConfigName("user")
		v.SetConfigFile(cfgFile)
		if err = v.MergeInConfig(); err != nil {
			fmt.Println("failed to merge user config with defaults, err:", err)
			os.Exit(1)
		}
	}

	v.AutomaticEnv() // read in environment variables that match

	err = v.Unmarshal(&cfg)
	if err != nil {
		fmt.Printf("unable to decode into config struct, %v", err)
	}
}
