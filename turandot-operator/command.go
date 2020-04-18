package main

import (
	"time"

	"github.com/spf13/cobra"
	puccinicommon "github.com/tliron/puccini/common"
	"github.com/tliron/puccini/common/terminal"
	"github.com/tliron/turandot/common"
)

var logTo string
var verbose int
var colorize bool

var masterUrl string
var kubeconfigPath string

var version bool
var site string
var cluster bool
var namespace string
var threadiness uint
var resyncPeriod time.Duration
var cachePath string
var healthPort uint

func init() {
	command.PersistentFlags().StringVarP(&logTo, "log", "l", "", "log to file (defaults to stderr)")
	command.PersistentFlags().CountVarP(&verbose, "verbose", "v", "add a log verbosity level (can be used twice)")
	command.PersistentFlags().BoolVarP(&colorize, "colorize", "z", true, "colorize output")

	// Conventional flags for Kubernetes controllers
	command.PersistentFlags().StringVar(&masterUrl, "master", "", "address of Kubernetes API server")
	command.PersistentFlags().StringVar(&kubeconfigPath, "kubeconfig", "", "path to Kubernetes configuration")

	// Our additional flags
	command.PersistentFlags().BoolVar(&version, "version", false, "print version")
	command.PersistentFlags().StringVar(&site, "site", "default", "site name")
	command.PersistentFlags().BoolVar(&cluster, "cluster", false, "enable cluster mode")
	command.PersistentFlags().StringVar(&namespace, "namespace", "", "namespace (overrides context namespace in Kubernetes configuration)")
	command.PersistentFlags().UintVar(&threadiness, "threadiness", 1, "number of concurrent workers per processor")
	command.PersistentFlags().DurationVar(&resyncPeriod, "resync", time.Second*30, "informer resync period")
	command.PersistentFlags().StringVar(&cachePath, "cache", "", "cache path")
	command.PersistentFlags().UintVar(&healthPort, "health-port", 8086, "HTTP port for health check (for liveness and readiness probes)")

	common.SetCobraFlagsFromEnvironment("TURANDOT_OPERATOR_", command)
}

var command = &cobra.Command{
	Use:   toolName,
	Short: "Start the Turandot operator",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if colorize {
			terminal.EnableColor()
		}
		if logTo == "" {
			puccinicommon.ConfigureLogging(verbose, nil)
		} else {
			puccinicommon.ConfigureLogging(verbose, &logTo)
		}
		// TODO: init "k8s.io/klog"?
	},
	Run: func(cmd *cobra.Command, args []string) {
		Operator()
	},
}

func Execute() {
	err := command.Execute()
	puccinicommon.FailOnError(err)
}
