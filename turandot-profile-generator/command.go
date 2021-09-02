package main

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

var command = &cobra.Command{
	Use:   "turandot-profile-generator [CONFIGURATION PATH]",
	Short: "Turandot Profile Generator",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		generator, err := NewGenerator(args[0])
		util.FailOnError(err)

		err = generator.Generate()
		util.FailOnError(err)
	},
}
