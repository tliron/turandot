package commands

import (
	"os/user"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tliron/puccini/common"
	"github.com/tliron/puccini/common/terminal"
)

var logTo string
var verbose int
var colorize bool
var kubeconfigPath string
var masterUrl string
var cluster bool
var namespace string

func init() {
	var defaultKubeconfigPath string
	if u, err := user.Current(); err == nil {
		defaultKubeconfigPath = filepath.Join(u.HomeDir, ".kube", "config")
	}

	rootCommand.PersistentFlags().StringVarP(&logTo, "log", "l", "", "log to file (defaults to stderr)")
	rootCommand.PersistentFlags().CountVarP(&verbose, "verbose", "v", "add a log verbosity level (can be used twice)")
	rootCommand.PersistentFlags().BoolVarP(&colorize, "colorize", "z", true, "colorize output")
	rootCommand.PersistentFlags().StringVarP(&masterUrl, "master", "m", "", "address of the Kubernetes API server")
	rootCommand.PersistentFlags().StringVarP(&kubeconfigPath, "kubeconfig", "k", defaultKubeconfigPath, "path to Kubernetes configuration")
	rootCommand.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace (overrides context namespace in Kubernetes configuration)")
}

var rootCommand = &cobra.Command{
	Use:   toolName,
	Short: "Control Turandot",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if colorize {
			terminal.EnableColor()
		}
		if logTo == "" {
			common.ConfigureLogging(verbose, nil)
		} else {
			common.ConfigureLogging(verbose, &logTo)
		}
	},
}

func Execute() {
	err := rootCommand.Execute()
	common.FailOnError(err)
}
