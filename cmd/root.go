package cmd

import (
	"fmt"
	"os"

	"github.com/tschokko/mdthklb/config"
	"github.com/spf13/cobra"
)

var (
	Version   = "dev-master"
	BuildTime = "undefined"
	GitHash   = "undefined"
)

var c = new(config.Config)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "mdthklb",
	Short: "Mediathek load balancer",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.UsageString())
		os.Exit(2)
	},
}

// Execute runs mdthklb and is called by main.main()
func Execute() {
	c.BuildTime = BuildTime
	c.BuildVersion = Version
	c.BuildHash = GitHash

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
