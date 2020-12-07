package commands

import (
	"os/user"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
)

var logTo string
var verbose int
var maxWidth int
var colorize string
var format string
var pretty bool
var strict bool

var masterUrl string
var kubeconfigPath string
var kubeconfigContext string
var clusterMode bool
var namespace string

func init() {
	var defaultKubeconfigPath string
	if u, err := user.Current(); err == nil {
		defaultKubeconfigPath = filepath.Join(u.HomeDir, ".kube", "config")
	}

	rootCommand.PersistentFlags().StringVarP(&logTo, "log", "l", "", "log to file (defaults to stderr)")
	rootCommand.PersistentFlags().CountVarP(&verbose, "verbose", "v", "add a log verbosity level (can be used twice)")
	rootCommand.PersistentFlags().IntVarP(&maxWidth, "width", "j", 0, "maximum output width (0 to use terminal width, -1 for no maximum)")
	rootCommand.PersistentFlags().StringVarP(&colorize, "colorize", "z", "true", "colorize output (boolean or \"force\")")
	rootCommand.PersistentFlags().StringVarP(&format, "format", "o", "", "output format (\"bare\", \"yaml\", \"json\", or \"xml\")")
	rootCommand.PersistentFlags().BoolVarP(&strict, "strict", "y", false, "strict output (for \"YAML\" format only)")
	rootCommand.PersistentFlags().BoolVarP(&pretty, "pretty", "p", true, "prettify output")

	rootCommand.PersistentFlags().StringVarP(&masterUrl, "master", "m", "", "address of the Kubernetes API server")
	rootCommand.PersistentFlags().StringVarP(&kubeconfigPath, "kubeconfig", "k", defaultKubeconfigPath, "path to Kubernetes configuration")
	rootCommand.PersistentFlags().StringVarP(&kubeconfigContext, "context", "x", "", "name of context in Kubernetes configuration")
	rootCommand.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace (overrides context namespace in Kubernetes configuration)")
}

var rootCommand = &cobra.Command{
	Use:   toolName,
	Short: "Control Turandot",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := terminal.ProcessColorizeFlag(colorize)
		util.FailOnError(err)
		if logTo == "" {
			util.ConfigureLogging(verbose, nil)
		} else {
			util.ConfigureLogging(verbose, &logTo)
		}
	},
}

func Execute() {
	err := rootCommand.Execute()
	util.FailOnError(err)
}
