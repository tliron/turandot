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
	inventoryCommand.AddCommand(inventoryDescribeCommand)
}

var inventoryDescribeCommand = &cobra.Command{
	Use:   "describe [INVENTORY NAME]",
	Short: "Describe an inventory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		DescribeInventory(args[0])
	},
}

func DescribeInventory(inventoryName string) {
	// TODO: in cluster mode we must specify the namespace
	namespace := ""

	inventory, err := NewClient().Turandot().GetInventory(namespace, inventoryName)
	util.FailOnError(err)

	if format != "" {
		formatpkg.Print(InventoryToARD(inventory), format, terminal.Stdout, strict, pretty)
	} else {
		fmt.Fprintf(terminal.Stdout, "%s: %s\n", terminal.ColorTypeName("Name"), terminal.ColorValue(inventory.Name))
		fmt.Fprintf(terminal.Stdout, "%s: %s\n", terminal.ColorTypeName("URL"), terminal.ColorValue(inventory.Spec.URL))
		fmt.Fprintf(terminal.Stdout, "%s: %s\n", terminal.ColorTypeName("SpoolerPod"), terminal.ColorValue(inventory.Status.SpoolerPod))
	}
}

func InventoryToARD(inventory *resources.Inventory) ard.StringMap {
	map_ := make(ard.StringMap)
	map_["Name"] = inventory.Name
	map_["URL"] = inventory.Spec.URL
	map_["SpoolerPod"] = inventory.Status.SpoolerPod
	return map_
}
