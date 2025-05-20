package viewgather

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/errors"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

var (
	commandExample = `
kubectl view gather --must-gather-path=~/my-must-gather get pods --namespace my-namespace
kubectl view gather --must-gather-path=~/my-must-gather describe node my-node
kubectl view gather --must-gather-path=~/my-must-gather logs my-pod
`
)

const (
	kubeconfigMustGatherName = "must-gather"
)

// ViewGatherOptions provides information required to update
// the current context on a user's KUBECONFIG
type ViewGatherOptions struct {
	configFlags *genericclioptions.ConfigFlags

	rawConfig api.Config
	args      []string

	mustGatherPath string

	genericiooptions.IOStreams
}

// NewViewGatherOptions provides an instance of ViewGatherOptions with default values
func NewViewGatherOptions(streams genericiooptions.IOStreams) *ViewGatherOptions {
	return &ViewGatherOptions{
		configFlags: genericclioptions.NewConfigFlags(true),

		IOStreams: streams,
	}
}

func NewViewGatherCommand(streams genericiooptions.IOStreams) *cobra.Command {
	o := NewViewGatherOptions(streams)

	cmd := &cobra.Command{
		Use:          "view gather [flags]",
		Short:        "Browse must-gather archive using kubectl-like commands",
		Example:      commandExample,
		SilenceUsage: true,
		Annotations: map[string]string{
			cobra.CommandDisplayNameAnnotation: "kubectl view gather",
		},
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			if err := o.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&o.mustGatherPath, "must-gather-path", o.mustGatherPath, "", "path to must-gather directory")
	o.configFlags.AddFlags(cmd.Flags())

	return cmd
}

func (o *ViewGatherOptions) Complete(cmd *cobra.Command, args []string) error {
	o.args = args

	var err error
	o.rawConfig, err = o.configFlags.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return err
	}

	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *ViewGatherOptions) Validate() error {
	var errs []error

	if len(o.mustGatherPath) == 0 {
		errs = append(errs, fmt.Errorf("must-gather-path cannot be empty"))
	}

	return errors.NewAggregate(errs)
}

func (o *ViewGatherOptions) Run() error {

	// TODO: spawn server faking kube-apiserver

	err := o.addNewMustGatherContextToKubeconfig("TODO Server address")
	if err != nil {
		return fmt.Errorf("can't save must gather context in kubeconfig: %w", err)
	}

	// TODO: execute kubectl with passed arguments and custom context

	return nil
}

func (o *ViewGatherOptions) addNewMustGatherContextToKubeconfig(server string) error {
	configAccess := clientcmd.NewDefaultPathOptions()

	resultingContext := api.NewContext()
	resultingContext.Cluster = kubeconfigMustGatherName
	resultingContext.AuthInfo = kubeconfigMustGatherName

	o.rawConfig.Contexts[kubeconfigMustGatherName] = resultingContext
	o.rawConfig.Clusters[kubeconfigMustGatherName] = &api.Cluster{
		Server:                "",
		TLSServerName:         "",
		InsecureSkipTLSVerify: false,
	}
	o.rawConfig.AuthInfos[kubeconfigMustGatherName] = &api.AuthInfo{}

	err := clientcmd.ModifyConfig(configAccess, o.rawConfig, true)
	if err != nil {
		return fmt.Errorf("can't save kubeconfig: %w", err)
	}

	return nil
}
