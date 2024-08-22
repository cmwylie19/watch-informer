package cmd

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/cmwylie19/watch-informer/pkg/logging"
	"github.com/cmwylie19/watch-informer/pkg/server"

	"github.com/spf13/cobra"
)

var logLevel string

var rootCmd = &cobra.Command{
	Use:   "watch-informer",
	Short: "Starts the watch-informer gRPC server",
	Run: func(cmd *cobra.Command, args []string) {
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

		server.StartGRPCServer(":50051", logger)
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
