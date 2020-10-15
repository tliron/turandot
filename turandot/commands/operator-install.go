package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	operatorCommand.AddCommand(operatorInstallCommand)
	operatorInstallCommand.Flags().StringVarP(&site, "site", "s", "default", "site name")
	operatorInstallCommand.Flags().BoolVarP(&cluster, "cluster", "c", false, "cluster mode")
	operatorInstallCommand.Flags().StringVarP(&registry, "registry", "g", "docker.io", "registry URL (use special value \"internal\" to discover internally deployed registry)")
	operatorInstallCommand.Flags().BoolVarP(&wait, "wait", "w", false, "wait for installation to succeed")
}

var operatorInstallCommand = &cobra.Command{
	Use:   "install",
	Short: "Install the Turandot operator",
	Run: func(cmd *cobra.Command, args []string) {
		err := NewClient().Turandot().InstallOperator(site, registry, wait)
		util.FailOnError(err)
	},
}
