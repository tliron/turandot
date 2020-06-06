package commands

import (
	"github.com/spf13/cobra"
	puccinicommon "github.com/tliron/puccini/common"
)

var site string
var registry string

func init() {
	rootCommand.AddCommand(installCommand)
	installCommand.PersistentFlags().StringVarP(&site, "site", "s", "default", "site name")
	installCommand.PersistentFlags().BoolVarP(&cluster, "cluster", "c", false, "cluster mode")
	installCommand.PersistentFlags().StringVarP(&registry, "registry", "r", "docker.io", "registry URL (use special value \"internal\" to discover internally deployed registry)")
	installCommand.PersistentFlags().BoolVarP(&wait, "wait", "w", false, "wait for installation to succeed")
}

var installCommand = &cobra.Command{
	Use:   "install",
	Short: "Install Turandot",
	Run: func(cmd *cobra.Command, args []string) {
		err := NewClient().Turandot().Install(site, registry, wait)
		puccinicommon.FailOnError(err)
	},
}
