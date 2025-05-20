package main

import (
	"os"

	"github.com/spf13/pflag"

	"github.com/zimnx/kubectl-view-gather/pkg/cmd/kubectl-view-gather"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func main() {
	flags := pflag.NewFlagSet("kubectl-view-gather", pflag.ExitOnError)
	pflag.CommandLine = flags

	root := viewgather.NewViewGatherCommand(genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
