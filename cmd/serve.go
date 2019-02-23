package cmd

import (
	"github.com/tschokko/mdthklb/cmd/server"
	"github.com/spf13/cobra"
)

// servePublicCmd represents the serve public command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serves the Mediathek load balancer application",
	Run:   server.RunServe(c),
}

func init() {
	RootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVarP(&c.ConfigFilename, "config", "c", "mdthklb.json", "The config file")
}
