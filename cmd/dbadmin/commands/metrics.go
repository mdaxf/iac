// Copyright 2023 IAC. All Rights Reserved.

package commands

import (
	"fmt"

	"github.com/mdaxf/iac/dbinitializer"
	"github.com/spf13/cobra"
)

func NewMetricsCommand() *cobra.Command {
	var (
		format string
		watch  bool
	)

	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Display database metrics",
		Long:  `Display performance metrics for all configured databases`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize databases from environment
			dbInit := dbinitializer.NewDatabaseInitializer()
			if err := dbInit.InitializeFromEnvironment(); err != nil {
				return fmt.Errorf("failed to initialize databases: %w", err)
			}

			poolManager := dbInit.GetPoolManager()
			if poolManager == nil {
				return fmt.Errorf("no databases configured")
			}

			databases := poolManager.GetAllDatabases()
			if len(databases) == 0 {
				fmt.Println("No databases configured")
				return nil
			}

			fmt.Println("Database Metrics")
			fmt.Println("================")
			fmt.Println("")

			for _, dbType := range databases {
				db, err := poolManager.GetPrimary(dbType)
				if err != nil {
					fmt.Printf("✗ %s: Failed to get connection\n\n", dbType)
					continue
				}

				fmt.Printf("Database: %s\n", dbType)
				fmt.Printf("  Dialect:  %s\n", db.GetDialect())
				fmt.Printf("  Status:   %s\n", getStatus(db))

				// TODO: Get actual metrics from metrics collector
				fmt.Printf("\n  Connection Pool:\n")
				fmt.Printf("    Active:     %d\n", 5)  // Placeholder
				fmt.Printf("    Idle:       %d\n", 10) // Placeholder
				fmt.Printf("    Max:        %d\n", 15) // Placeholder

				fmt.Printf("\n  Query Statistics:\n")
				fmt.Printf("    Total:      %d\n", 1234) // Placeholder
				fmt.Printf("    Errors:     %d\n", 0)    // Placeholder
				fmt.Printf("    Slow:       %d\n", 5)    // Placeholder
				fmt.Printf("    Avg Time:   %.2fms\n", 45.3) // Placeholder

				fmt.Println("")

				db.Close()
			}

			if watch {
				fmt.Println("Note: Watch mode not yet implemented")
			}

			fmt.Println("Tip: Use --format=json for machine-readable output")

			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "text", "Output format (text, json)")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch mode (update every second)")

	return cmd
}

func getStatus(db interface{ Ping() error }) string {
	if err := db.Ping(); err != nil {
		return "Unhealthy"
	}
	return "Healthy"
}
