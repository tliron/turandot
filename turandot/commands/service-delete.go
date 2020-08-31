package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	serviceCommand.AddCommand(serviceDeleteCommand)
	serviceDeleteCommand.Flags().BoolVarP(&all, "all", "a", false, "delete all services")
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
	// TODO: in cluster mode we must specify the namespace
	namespace := ""

	err := NewClient().Turandot().DeleteService(namespace, serviceName)
	util.FailOnError(err)
}

func DeleteAllServices() {
	turandot := NewClient().Turandot()
	services, err := turandot.ListServices()
	util.FailOnError(err)
	for _, service := range services.Items {
		log.Infof("deleting service: %s/%s", service.Namespace, service.Name)
		err := turandot.DeleteService(service.Namespace, service.Name)
		util.FailOnError(err)
	}
}
