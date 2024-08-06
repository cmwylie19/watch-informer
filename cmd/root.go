package cmd

import (
	"log"
	"watch-informer/server"

	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	group, version, resource, namespace string
)

var rootCmd = &cobra.Command{
	Use:   "watch-informer",
	Short: "Starts the watch-informer gRPC server",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			log.Fatalf("Error building kubeconfig: %s", err)
		}
		dynamicClient, err := dynamic.NewForConfig(config)
		if err != nil {
			log.Fatalf("Error creating dynamic client: %s", err)
		}
		server.StartGRPCServer(":50051", dynamicClient, group, version, resource, namespace)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&group, "group", "g", "", "Group of the Kubernetes resource")
	rootCmd.PersistentFlags().StringVarP(&version, "version", "v", "", "Version of the Kubernetes resource")
	rootCmd.PersistentFlags().StringVarP(&resource, "resource", "r", "", "Resource to watch")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Namespace to watch (empty for all namespaces)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("CLI execution error: %v", err)
	}
}
