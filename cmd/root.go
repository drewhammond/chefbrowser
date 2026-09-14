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
	rootCmd.PersistentFlags().String("listen-addr", "", "address to listen on (default: 0.0.0.0:8080)")
	rootCmd.PersistentFlags().String("app-mode", "", "application mode: production or development")
	rootCmd.PersistentFlags().Bool("use-mock-data", false, "use mock data instead of real Chef server")
	rootCmd.PersistentFlags().String("chef-server-url", "", "Chef server URL")
	rootCmd.PersistentFlags().String("chef-username", "", "Chef server username")
	rootCmd.PersistentFlags().String("chef-key-file", "", "path to Chef client key file")
	rootCmd.PersistentFlags().Bool("chef-ssl-verify", true, "verify Chef server SSL certificate")
	rootCmd.PersistentFlags().String("log-level", "", "log level: debug, info, warning, error, fatal")
	rootCmd.PersistentFlags().String("log-format", "", "log format: json or console")
	rootCmd.PersistentFlags().String("log-output", "", "log output: stdout or file path")
	rootCmd.PersistentFlags().Bool("request-logging", true, "enable request logging")
	rootCmd.PersistentFlags().Bool("log-health-checks", true, "log health check requests")
	rootCmd.PersistentFlags().String("base-path", "", "base path for reverse proxy")
	rootCmd.PersistentFlags().String("trusted-proxies", "", "comma-separated trusted proxy CIDRs")
	rootCmd.PersistentFlags().Bool("enable-gzip", false, "enable gzip compression")
}

// initConfig reads in config defaults, user config files, and ENV variables if set.
func initConfig() {
	// ignore inline comments to allow # characters in the middle of custom links and other properties - (#399)
	v := viper.NewWithOptions(viper.IniLoadOptions(ini.LoadOptions{IgnoreInlineComment: true}))
	v.SetConfigType("ini")
	v.SetConfigName("chefbrowser")
	v.AddConfigPath("/etc/chefbrowser/")

	v.BindPFlag("default.listen_addr", rootCmd.PersistentFlags().Lookup("listen-addr"))
	v.BindPFlag("default.app_mode", rootCmd.PersistentFlags().Lookup("app-mode"))
	v.BindPFlag("default.use_mock_data", rootCmd.PersistentFlags().Lookup("use-mock-data"))
	v.BindPFlag("chef.server_url", rootCmd.PersistentFlags().Lookup("chef-server-url"))
	v.BindPFlag("chef.username", rootCmd.PersistentFlags().Lookup("chef-username"))
	v.BindPFlag("chef.key_file", rootCmd.PersistentFlags().Lookup("chef-key-file"))
	v.BindPFlag("chef.ssl_verify", rootCmd.PersistentFlags().Lookup("chef-ssl-verify"))
	v.BindPFlag("logging.level", rootCmd.PersistentFlags().Lookup("log-level"))
	v.BindPFlag("logging.format", rootCmd.PersistentFlags().Lookup("log-format"))
	v.BindPFlag("logging.output", rootCmd.PersistentFlags().Lookup("log-output"))
	v.BindPFlag("logging.request_logging", rootCmd.PersistentFlags().Lookup("request-logging"))
	v.BindPFlag("logging.log_health_checks", rootCmd.PersistentFlags().Lookup("log-health-checks"))
	v.BindPFlag("server.base_path", rootCmd.PersistentFlags().Lookup("base-path"))
	v.BindPFlag("server.trusted_proxies", rootCmd.PersistentFlags().Lookup("trusted-proxies"))
	v.BindPFlag("server.enable_gzip", rootCmd.PersistentFlags().Lookup("enable-gzip"))

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
