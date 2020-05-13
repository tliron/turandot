package commands

import (
	"github.com/spf13/cobra"
	puccinicommon "github.com/tliron/puccini/common"
)

func init() {
	serviceCommand.AddCommand(serviceDeleteCommand)
	serviceDeleteCommand.PersistentFlags().BoolVarP(&all, "all", "a", false, "delete all services")
}

var serviceDeleteCommand = &cobra.Command{
	Use:   "delete [[SERVICE NAME]]",
	Short: "Delete a deployed service",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			serviceName := args[0]
			DeleteService(serviceName)
		} else if all {
			DeleteAllServices()
		}
	},
}

func DeleteService(serviceName string) {
	err := NewClient().Turandot().DeleteService(serviceName)
	puccinicommon.FailOnError(err)
}

func DeleteAllServices() {
	turandot := NewClient().Turandot()
	services, err := turandot.ListServices()
	puccinicommon.FailOnError(err)
	for _, service := range services.Items {
		log.Infof("deleting service: %s", service.Name)
		err := turandot.DeleteService(service.Name)
		puccinicommon.FailOnError(err)
	}
}
