package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	verbose bool
	RootCmd = &cobra.Command{
		Use:   "im-concierge",
		Short: "IM concierge daemon",
		Long:  `IM concierge daemon connects to IM servers and expose HTTP interface to access your discussions`,
		Run:   nil,
	}
)

func init() {
	cobra.OnInitialize()
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
		"print out more debug information")
	RootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if verbose {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.InfoLevel)
		}
	}
}
