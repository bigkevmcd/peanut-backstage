package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-logr/zapr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/bigkevmcd/peanut-backstage/pkg/httpapi"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(appsv1.AddToScheme(scheme))
	cobra.OnInitialize(initConfig)
}

const (
	listenFlag             = "listen"
	debugFlag              = "debug"
	shutdownTimeoutSeconds = 30
)

func initConfig() {
	viper.AutomaticEnv()
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "peanut-backstage",
		Short: "Export Kubernetes resources as a Backstage catalog",
	}

	cmd.AddCommand(newServeCmd())

	return cmd
}

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Dynamic HTTP server serving Backstage components",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetConfig()
			cobra.CheckErr(err)

			cl, err := client.New(cfg, client.Options{Scheme: scheme})
			cobra.CheckErr(err)

			logger := zapr.NewLogger(makeLogger(viper.GetBool(debugFlag)))
			router := httpapi.NewRouter(logger, cl)

			listen := viper.GetString(listenFlag)
			
			server := &http.Server{
				Addr:              listen,
				Handler:           router,
				ReadHeaderTimeout: 10 * time.Second,
				ReadTimeout:       30 * time.Second,
				WriteTimeout:      30 * time.Second,
				IdleTimeout:       120 * time.Second,
			}

			// Channel to listen for errors coming from the listener.
			serverErrors := make(chan error, 1)

			// Start the server
			go func() {
				logger.Info("starting server", "addr", listen)
				fmt.Printf("serving the root catalog at http://%s/backstage/catalog-info.yaml\n", listen)
				serverErrors <- server.ListenAndServe()
			}()

			// Channel to listen for an interrupt or terminate signal from the OS.
			shutdown := make(chan os.Signal, 1)
			signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

			// Blocking main and waiting for shutdown.
			select {
			case err := <-serverErrors:
				return fmt.Errorf("server error: %w", err)

			case sig := <-shutdown:
				logger.Info("shutdown signal received", "signal", sig)

				// Give outstanding requests a deadline for completion.
				ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeoutSeconds*time.Second)
				defer cancel()

				// Asking listener to shut down and shed load.
				if err := server.Shutdown(ctx); err != nil {
					server.Close()
					return fmt.Errorf("could not stop server gracefully: %w", err)
				}
				
				logger.Info("server stopped gracefully")
			}

			return nil
		},
	}

	cmd.Flags().String(
		listenFlag,
		"localhost:8080",
		"listen address e.g. :8080",
	)
	cmd.Flags().Bool(
		debugFlag,
		false,
		"enable debug logging",
	)
	cobra.CheckErr(viper.BindPFlag(listenFlag, cmd.Flags().Lookup(listenFlag)))
	return cmd
}

// Execute is the main entry point into this component.
func Execute() {
	cobra.CheckErr(newRootCmd().Execute())
}

func makeLogger(debug bool) *zap.Logger {
	var zapLog *zap.Logger
	var err error
	if debug {
		zapLog, err = zap.NewDevelopment()
	} else {
		zapLog, err = zap.NewProduction()
	}
	cobra.CheckErr(err)
	return zapLog
}
