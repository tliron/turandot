package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	inventoryCommand.AddCommand(inventoryDeleteCommand)
	inventoryDeleteCommand.Flags().BoolVarP(&all, "all", "a", false, "delete all inventories")
}

var inventoryDeleteCommand = &cobra.Command{
	Use:   "delete [[INVENTORY NAME]]",
	Short: "Delete an inventory",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			inventoryName := args[0]
			DeleteInventory(inventoryName)
		} else if all {
			DeleteAllInventories()
		} else {
			util.Fail("must provide inventory name or specify \"--all\"")
		}
	},
}

func DeleteInventory(inventoryName string) {
	// TODO: in cluster mode we must specify the namespace
	namespace := ""

	err := NewClient().Turandot().DeleteInventory(namespace, inventoryName)
	util.FailOnError(err)
}

func DeleteAllInventories() {
	turandot := NewClient().Turandot()
	inventories, err := turandot.ListInventories()
	util.FailOnError(err)
	for _, inventory := range inventories.Items {
		log.Infof("deleting inventory: %s/%s", inventory.Namespace, inventory.Name)
		err := turandot.DeleteInventory(inventory.Namespace, inventory.Name)
		util.FailOnError(err)
	}
}
