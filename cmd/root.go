package cmd

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/cmwylie19/watch-informer/pkg/logging"
	"github.com/cmwylie19/watch-informer/pkg/server"

	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var logLevel string
var useInClusterConfig bool

var (
	getInClusterConfig     = rest.InClusterConfig
	getConfigFromFlags     = clientcmd.BuildConfigFromFlags
	getDynamicNewForConfig = dynamic.NewForConfig
	createLogger           = logging.NewLogger
)

var rootCmd = &cobra.Command{
	Use:   "watch-informer",
	Short: "Starts the watch-informer gRPC server",
	Run: func(cmd *cobra.Command, args []string) {
		var config *rest.Config
		var err error

		if useInClusterConfig {
			config, err = getInClusterConfig()
			if err != nil {
				log.Fatalf("Error building in-cluster config: %s", err)
			}
		} else {
			config, err = getConfigFromFlags("", clientcmd.RecommendedHomeFile)
			if err != nil {
				log.Fatalf("Error building kubeconfig: %s", err)
			}
		}

		dynamicClient, err := getDynamicNewForConfig(config)
		if err != nil {
			log.Fatalf("Error creating dynamic client: %s", err)
		}

		logger, err := createLogger("")
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
	rootCmd.PersistentFlags().BoolVar(&useInClusterConfig, "in-cluster", true, "Use in-cluster configuration")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("CLI execution error: %v", err)
	}
}
