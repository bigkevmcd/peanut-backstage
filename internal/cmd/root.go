package cmd

import (
	"fmt"
	"log"
	"net/http"

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
	portFlag  = "port"
	debugFlag = "debug"
)

func initConfig() {
	viper.AutomaticEnv()
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "peanut-backstage",
		Short: "Export Kubernetes resources as a Backstage catalog",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetConfig()
			cobra.CheckErr(err)

			cl, err := client.New(cfg, client.Options{Scheme: scheme})
			cobra.CheckErr(err)

			logger := zapr.NewLogger(makeLogger(viper.GetBool(debugFlag)))
			router := httpapi.NewRouter(logger, cl)

			port := viper.GetString(portFlag)
			log.Printf("listening on http://localhost:%s/", port)
			return http.ListenAndServe(fmt.Sprintf(":%s", port), router)
		},
	}

	cmd.Flags().Int(
		portFlag,
		8080,
		"port to serve requests on",
	)
	cmd.Flags().Bool(
		debugFlag,
		false,
		"enable debug logging",
	)
	cobra.CheckErr(viper.BindPFlag(portFlag, cmd.Flags().Lookup(portFlag)))
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
