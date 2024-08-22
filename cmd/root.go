package cmd

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"watch-informer/pkg/logging"
	"watch-informer/pkg/server"

	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

var logLevel string
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
		logger, err := logging.NewLogger("")
		if err != nil {
			fmt.Printf("Failed to initialize logger: %v\n", err)
			os.Exit(1)
		}
		defer logger.CloseFile()

		switch logLevel {
		case "debug":
			logger.SetLevel(slog.LevelDebug)
		case "info":
			logger.SetLevel(slog.LevelInfo)
		case "warn":
			logger.SetLevel(slog.LevelWarn)
		case "error":
			logger.SetLevel(slog.LevelError)
		default:
			logger.SetLevel(slog.LevelInfo) // Default to INFO level
		}

		server.StartGRPCServer(":50051", dynamicClient, logger)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (debug, info, error)")
}
func Execute() {

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("CLI execution error: %v", err)
	}
}
