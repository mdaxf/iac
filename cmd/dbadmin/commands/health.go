// Copyright 2023 IAC. All Rights Reserved.

package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mdaxf/iac/dbinitializer"
	"github.com/spf13/cobra"
)

func NewHealthCommand() *cobra.Command {
	var (
		verbose bool
		json    bool
	)

	cmd := &cobra.Command{
		Use:   "health",
		Short: "Check database health",
		Long:  `Check health status of all configured databases`,
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

			fmt.Printf("Checking health of %d database(s)...\n\n", len(databases))

			healthyCount := 0
			unhealthyCount := 0

			for _, dbType := range databases {
				db, err := poolManager.GetPrimary(dbType)
				if err != nil {
					fmt.Printf("✗ %s: Failed to get connection - %v\n", dbType, err)
					unhealthyCount++
					continue
				}

				start := time.Now()
				err = db.Ping()
				duration := time.Since(start)

				if err != nil {
					fmt.Printf("✗ %s: Unhealthy - %v\n", dbType, err)
					unhealthyCount++
				} else {
					fmt.Printf("✓ %s: Healthy (ping: %v)\n", dbType, duration)
					healthyCount++

					if verbose {
						dialect := db.GetDialect()
						fmt.Printf("    Dialect: %s\n", dialect)
					}
				}

				db.Close()
			}

			fmt.Printf("\nSummary:\n")
			fmt.Printf("  Healthy:   %d\n", healthyCount)
			fmt.Printf("  Unhealthy: %d\n", unhealthyCount)
			fmt.Printf("  Total:     %d\n", len(databases))

			if unhealthyCount > 0 {
				os.Exit(1)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")
	cmd.Flags().BoolVarP(&json, "json", "j", false, "Output in JSON format")

	return cmd
}
