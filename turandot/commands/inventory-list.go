package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tliron/kutil/ard"
	formatpkg "github.com/tliron/kutil/format"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
)

func init() {
	inventoryCommand.AddCommand(inventoryListCommand)
}

var inventoryListCommand = &cobra.Command{
	Use:   "list",
	Short: "List inventories",
	Run: func(cmd *cobra.Command, args []string) {
		ListInventories()
	},
}

func ListInventories() {
	inventories, err := NewClient().Turandot().ListInventories()
	util.FailOnError(err)
	if len(inventories.Items) == 0 {
		return
	}
	// TODO: sort inventories by name? they seem already sorted!

	switch format {
	case "":
		table := terminal.NewTable(maxWidth, "Name", "URL", "SpoolerPod")
		for _, inventory := range inventories.Items {
			table.Add(inventory.Name, inventory.Spec.URL, inventory.Status.SpoolerPod)
		}
		table.Print()

	case "bare":
		for _, inventory := range inventories.Items {
			fmt.Fprintln(terminal.Stdout, inventory.Name)
		}

	default:
		list := make(ard.List, len(inventories.Items))
		for index, inventory := range inventories.Items {
			list[index] = InventoryToARD(&inventory)
		}
		formatpkg.Print(list, format, terminal.Stdout, strict, pretty)
	}
}
