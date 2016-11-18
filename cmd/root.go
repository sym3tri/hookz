package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	log             *logrus.Logger
	defaultLogLevel = logrus.InfoLevel
	rootCmdFile     string
	rootCmdCfg      *viper.Viper
)

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "hookz",
	Short: "A webhook hanlder for scheduling kubernetes jobs",
	Long:  `More info at https://github.com/sym3tri/hookz`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		rootCmdCfg.BindPFlags(cmd.Flags())
		initRootLogger()
	},
}

func init() {
	log = logrus.New()
	cobra.OnInitialize(func() {
		rootCmdCfg = newViper("hookz", rootCmdFile)
	})
	// Persistent Flags, which are global for the application.
	RootCmd.PersistentFlags().StringVar(&rootCmdFile, "config", "", "Config file (default is $HOME/.hookz.yaml)")
	RootCmd.PersistentFlags().String("log-level", defaultLogLevel.String(), "Set the global log level")
}

func initRootLogger() {
	// root level logger
	log = logrus.New()
	// TODO: in next version of logrus
	//logrus.RegisterExitHandler(func() {
	//log.Info("bye bye")
	//})

	// Always attempt to configure log-level before any subcommands run.
	logLevel := parseLogLevel(rootCmdCfg.GetString("log-level"))
	log.Level = logLevel
	log.WithField("log-level", log.Level).Debug("root log level set")
}

func parseLogLevel(level string) logrus.Level {
	if level == "" {
		return defaultLogLevel
	}

	ll, err := logrus.ParseLevel(level)
	if err != nil {
		fmt.Printf("failed to parse log-level, using default. error: %v\n", err)
		return defaultLogLevel
	}
	return ll
}

func newLogger(app, level string) *logrus.Entry {
	logger := logrus.New()
	logger.Level = parseLogLevel(level)
	logEntry := logrus.NewEntry(logger).WithFields(logrus.Fields{
		"app": app,
	})
	logEntry.WithField("log-level", logger.Level).Info("log level set")
	return logEntry
}

func newViper(appName, cfgFile string) *viper.Viper {
	v := viper.New()
	v.SetEnvPrefix(appName)
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	// name of config file (without extension)
	v.SetConfigName(fmt.Sprintf("config-%s", appName))
	v.AddConfigPath(".")
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	}

	// If a config file is found, read it in.
	if err := v.ReadInConfig(); err == nil {
		log.WithField("file", v.ConfigFileUsed()).Infof("config file loaded")
		v.WatchConfig()
		v.OnConfigChange(func(e fsnotify.Event) {
			log.WithField("file", e.Name).Info("config file changed")
		})
	} else {
		if cfgFile != "" {
			log.WithError(err).Warningf("error loading config file: %s", cfgFile)
		}
	}

	if log.Level == logrus.DebugLevel {
		v.Debug()
	}

	return v
}
