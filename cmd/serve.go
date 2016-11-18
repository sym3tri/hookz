package cmd

import (
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sym3tri/hookz/server"
)

var (
	serveCmdCfg  *viper.Viper
	serveCmdFile string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the hook handler HTTP server.",
	Long:  `Define your handler routes and tasks in the config file`,
	PreRun: func(cmd *cobra.Command, args []string) {
		serveCmdCfg = newViper("serve", serveCmdFile)
		serveCmdCfg.BindPFlags(cmd.Flags())
	},
	Run: RunServer,
}

func init() {
	RootCmd.AddCommand(serveCmd)
	serveCmd.PersistentFlags().StringVar(&serveCmdFile, "config", "", "Config file for serve command")
	serveCmd.PersistentFlags().String("listen", "0.0.0.0:8080", "Host and port to listen on")
	serveCmd.PersistentFlags().String("tls-cert-file", "", "TLS certificate. If the certificate is signed by a certificate authority, the certFile should be the concatenation of the server's certificate followed by the CA's certificate.")
	serveCmd.PersistentFlags().String("tls-key-file", "", "The TLS certificate key.")
	serveCmd.PersistentFlags().String("log-level", defaultLogLevel.String(), "Set the global log level")
}

func RunServer(cmd *cobra.Command, args []string) {
	logger := newLogger("hookz-server", serveCmdCfg.GetString("log-level"))
	listen := serveCmdCfg.GetString("listen")
	logger.WithField("listen", listen).Info("listening...")

	logger.Debugf("raw endpoints: %v", serveCmdCfg.Get("endpoints"))
	var endpoints []server.Endpoint
	if err := serveCmdCfg.UnmarshalKey("endpoints", &endpoints); err != nil {
		logger.WithError(err).Fatal("unable to decode config values")
	}
	logger.Debugf("endponits: %+v", endpoints)

	srv := server.New(logger, endpoints)
	srv.Version = Version
	httpsvr := &http.Server{
		Addr:    listen,
		Handler: srv.HTTPHandler(),
	}

	tlsCertFile := serveCmdCfg.GetString("tls-cert-file")
	tlsKeyFile := serveCmdCfg.GetString("tls-key-file")
	if tlsCertFile != "" && tlsKeyFile == "" {
		log.Info("serving over https")
		log.Fatal(httpsvr.ListenAndServeTLS(tlsCertFile, tlsKeyFile))
	} else {
		log.Warning("serving over http")
		log.Fatal(httpsvr.ListenAndServe())
	}
}
