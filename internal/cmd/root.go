package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bigkevmcd/peanut-backstage/pkg/httpapi"
	"github.com/go-logr/glogr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	portFlag = "port"
)

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.AutomaticEnv()
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "peanut-backstage",
		Short: "Export Kubernetes resources as a Backstage catalog",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := glogr.New()

			router := httpapi.NewRouter(logger)

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
	cobra.CheckErr(viper.BindPFlag(portFlag, cmd.Flags().Lookup(portFlag)))
	return cmd
}

// Execute is the main entry point into this component.
func Execute() {
	cobra.CheckErr(newRootCmd().Execute())
}
