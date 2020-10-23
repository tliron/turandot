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
	repositoryCommand.AddCommand(repositoryDescribeCommand)
}

var repositoryDescribeCommand = &cobra.Command{
	Use:   "describe [INVENTORY NAME]",
	Short: "Describe a repository",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		DescribeRepository(args[0])
	},
}

func DescribeRepository(repositoryName string) {
	// TODO: in cluster mode we must specify the namespace
	namespace := ""

	repository, err := NewClient().Turandot().GetRepository(namespace, repositoryName)
	util.FailOnError(err)

	if format != "" {
		formatpkg.Print(RepositoryToARD(repository), format, terminal.Stdout, strict, pretty)
	} else {
		fmt.Fprintf(terminal.Stdout, "%s: %s\n", terminal.ColorTypeName("Name"), terminal.ColorValue(repository.Name))
		fmt.Fprintf(terminal.Stdout, "%s: %s\n", terminal.ColorTypeName("URL"), terminal.ColorValue(repository.Spec.URL))
		fmt.Fprintf(terminal.Stdout, "%s: %s\n", terminal.ColorTypeName("SpoolerPod"), terminal.ColorValue(repository.Status.SpoolerPod))
	}
}

func RepositoryToARD(repository *resources.Repository) ard.StringMap {
	map_ := make(ard.StringMap)
	map_["Name"] = repository.Name
	map_["URL"] = repository.Spec.URL
	map_["SpoolerPod"] = repository.Status.SpoolerPod
	return map_
}
