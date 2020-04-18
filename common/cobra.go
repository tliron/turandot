package common

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func SetCobraFlagsFromEnvironment(prefix string, command *cobra.Command) {
	setCobraFlagsFromEnvironment(prefix, command.PersistentFlags())
	setCobraFlagsFromEnvironment(prefix, command.Flags())
}

func setCobraFlagsFromEnvironment(prefix string, flags *pflag.FlagSet) {
	flags.VisitAll(func(flag *pflag.Flag) {
		if value, ok := os.LookupEnv(prefix + flag.Name); ok {
			flags.Set(flag.Name, value)
		}
	})
}
