package commands

import (
	"os/user"
	"path/filepath"

	"github.com/spf13/cobra"
	puccinicommon "github.com/tliron/puccini/common"
)

var delegateKubeconfigPath string
var delegateContext string
var delegateNamespace string

func init() {
	var defaultKubeconfigPath string
	if u, err := user.Current(); err == nil {
		defaultKubeconfigPath = filepath.Join(u.HomeDir, ".kube", "config")
	}

	delegateCommand.AddCommand(delegateSetCommand)
	delegateSetCommand.Flags().StringVarP(&delegateKubeconfigPath, "delegate-kubeconfig", "", defaultKubeconfigPath, "path to delegate delegate Kubernetes configuration")
	delegateSetCommand.Flags().StringVarP(&delegateContext, "delegate-context", "", "", "override current context in delegate Kubernetes configuration")
	delegateSetCommand.Flags().StringVarP(&delegateNamespace, "delegate-namespace", "", "", "namespace (overrides context namespace in delegate Kubernetes configuration)")
}

var delegateSetCommand = &cobra.Command{
	Use:   "set [DELEGATE NAME]",
	Short: "Update or create a delegate",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		delegateName := args[0]
		SetDelegate(delegateName)
	},
}

func SetDelegate(delegateName string) {
	err := NewClient().Turandot().SetDelegate(delegateName, delegateKubeconfigPath, delegateContext, delegateNamespace)
	puccinicommon.FailOnError(err)
}
