package viewgather

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zimnx/kubectl-view-gather/pkg/mustgather"
	"github.com/zimnx/kubectl-view-gather/pkg/server"
	"github.com/zimnx/kubectl-view-gather/pkg/store"
	"io"
	k8serrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"net/http"
	"os/exec"
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
		Use:          "view gather [flags] -- [kubectl command and flags]",
		Short:        "Browse must-gather archive using kubectl-like commands",
		Example:      commandExample,
		SilenceUsage: true,
		Annotations: map[string]string{
			cobra.CommandDisplayNameAnnotation: "kubectl view gather",
		},
		Args: cobra.ArbitraryArgs, // Accept arbitrary args after --
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

	return k8serrors.NewAggregate(errs)
}

func (o *ViewGatherOptions) Run() error {
	mg := mustgather.NewMustGatherArchive(o.mustGatherPath)
	objectStore := store.NewObject()
	apiServer := server.NewAPIServerStub(mg, objectStore)
	serverAddress := "localhost:8080" // TODO: make this configurable or pick a free port dynamically
	go func() {
		if err := http.ListenAndServe(serverAddress, apiServer); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				fmt.Printf("Failed to start server: %v", err)
			}
		}
	}()

	err := o.addNewMustGatherContextToKubeconfig(fmt.Sprintf("http://%s", serverAddress))
	if err != nil {
		return fmt.Errorf("can't save must gather context in kubeconfig: %w", err)
	}

	args := filterMustGatherArgs(o.args)
	args = append(args, "--context", kubeconfigMustGatherName)
	out, err := exec.Command("kubectl", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run kubectl: %w\n%s", err, string(out))
	}

	if _, err = io.WriteString(o.Out, string(out)); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

func (o *ViewGatherOptions) addNewMustGatherContextToKubeconfig(server string) error {
	configAccess := clientcmd.NewDefaultPathOptions()

	resultingContext := api.NewContext()
	resultingContext.Cluster = kubeconfigMustGatherName
	resultingContext.AuthInfo = kubeconfigMustGatherName

	o.rawConfig.Contexts[kubeconfigMustGatherName] = resultingContext
	o.rawConfig.Clusters[kubeconfigMustGatherName] = &api.Cluster{
		Server: server,
	}
	o.rawConfig.AuthInfos[kubeconfigMustGatherName] = &api.AuthInfo{}

	err := clientcmd.ModifyConfig(configAccess, o.rawConfig, true)
	if err != nil {
		return fmt.Errorf("can't save kubeconfig: %w", err)
	}

	return nil
}

func filterMustGatherArgs(args []string) []string {
	var filtered []string
	skipNext := false

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if skipNext {
			// Skip the value for --must-gather-path
			skipNext = false
			continue
		}

		if arg == "--must-gather-path" {
			// This is the split form; skip next arg too
			skipNext = true
			continue
		}

		if len(arg) > 20 && arg[:20] == "--must-gather-path=" {
			// This is the equals form; skip
			continue
		}

		filtered = append(filtered, arg)
	}

	return filtered
}
