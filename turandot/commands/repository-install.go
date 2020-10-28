package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

var secure bool

func init() {
	repositoryCommand.AddCommand(repositoryInstallCommand)
	repositoryInstallCommand.Flags().BoolVarP(&cluster, "cluster", "c", false, "cluster mode")
	repositoryInstallCommand.Flags().StringVarP(&registry, "registry", "g", "docker.io", "registry URL (use special value \"internal\" to discover internally deployed registry)")
	repositoryInstallCommand.Flags().BoolVarP(&secure, "secure", "s", true, "secure the repository (requires cert-manager)")
	repositoryInstallCommand.Flags().BoolVarP(&wait, "wait", "w", false, "wait for installation to succeed")
}

var repositoryInstallCommand = &cobra.Command{
	Use:   "install",
	Short: "Install the Turandot repository",
	Run: func(cmd *cobra.Command, args []string) {
		err := NewClient().Turandot().InstallRepository(registry, secure, wait)
		util.FailOnError(err)
	},
}
