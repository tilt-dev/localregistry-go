package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tilt-dev/localregistry-go"
	"gopkg.in/yaml.v2"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var cf *genericclioptions.ConfigFlags

var rootCmd = &cobra.Command{
	Use:   "local-registry [command]",
	Short: "Kubectl plugin for interacting with local registries",
	Example: "  kubectl local-registry get\n" +
		"  kubectl local-registry get --context=microk8s",
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the local registry advertised by the cluster",
	Example: "  kubectl local-registry get\n" +
		"  kubectl local-registry get --context=microk8s",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		core, err := core()
		if err != nil {
			exit(fmt.Errorf("Connecting to Kubernetes: %v", err))
		}
		hosting, err := localregistry.Discover(ctx, core)
		if err != nil {
			exit(fmt.Errorf("Detecting local registry: %v", err))
		}
		if hosting == (localregistry.LocalRegistryHostingV1{}) {
			fmt.Printf("Local registry not found in cluster\n")
			os.Exit(0)
		}

		err = yaml.NewEncoder(os.Stdout).Encode(hosting)
		if err != nil {
			exit(fmt.Errorf("Printing local registry: %v", err))
		}
	},
}

func main() {
	cf = genericclioptions.NewConfigFlags(true)
	cf.AddFlags(getCmd.Flags())

	rootCmd.AddCommand(getCmd)

	err := rootCmd.Execute()
	if err != nil {
		// Cobra already printed the error
		os.Exit(1)
	}
}

func core() (v1.CoreV1Interface, error) {
	config, err := cf.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		exit(err)
	}
	return clientset.CoreV1(), nil
}

func exit(err error) {
	fmt.Printf("Error: %v\n", err)
	os.Exit(1)
}
