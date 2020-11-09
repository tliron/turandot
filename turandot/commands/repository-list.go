package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tliron/kutil/ard"
	formatpkg "github.com/tliron/kutil/format"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
)

func init() {
	repositoryCommand.AddCommand(repositoryListCommand)
}

var repositoryListCommand = &cobra.Command{
	Use:   "list",
	Short: "List repositories",
	Run: func(cmd *cobra.Command, args []string) {
		ListRepositories()
	},
}

func ListRepositories() {
	repositories, err := NewClient().Turandot().ListRepositories()
	util.FailOnError(err)
	if len(repositories.Items) == 0 {
		return
	}
	// TODO: sort repositories by name? they seem already sorted!

	switch format {
	case "":
		table := terminal.NewTable(maxWidth, "Name", "Host", "Namespace", "Service", "Port", "SpoolerPod")
		for _, repository := range repositories.Items {
			if repository.Spec.Direct != nil {
				table.Add(repository.Name, repository.Spec.Direct.Host, "", "", "", repository.Status.SpoolerPod)
			} else if repository.Spec.Indirect != nil {
				table.Add(repository.Name, "", repository.Spec.Indirect.Namespace, repository.Spec.Indirect.Service, fmt.Sprintf("%d", repository.Spec.Indirect.Port), repository.Status.SpoolerPod)
			}
		}
		table.Print()

	case "bare":
		for _, repository := range repositories.Items {
			fmt.Fprintln(terminal.Stdout, repository.Name)
		}

	default:
		list := make(ard.List, len(repositories.Items))
		for index, repository := range repositories.Items {
			list[index] = resources.RepositoryToARD(&repository)
		}
		formatpkg.Print(list, format, terminal.Stdout, strict, pretty)
	}
}
