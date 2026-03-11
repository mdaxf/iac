// Copyright 2023 IAC. All Rights Reserved.

package commands

import (
	"fmt"
	"os"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
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

			poolManager := dbconn.GetPoolManager()
			if poolManager == nil {
				return fmt.Errorf("no databases configured")
			}

			healthyCount := 0
			unhealthyCount := 0

			checkDB := func(label string, db dbconn.RelationalDB) {
				start := time.Now()
				err := db.Ping()
				duration := time.Since(start)

				if err != nil {
					fmt.Printf("✗ %s: Unhealthy - %v\n", label, err)
					unhealthyCount++
				} else {
					fmt.Printf("✓ %s: Healthy (ping: %v)\n", label, duration)
					healthyCount++

					if verbose {
						fmt.Printf("    Dialect: %s\n", db.GetDialect())
					}
				}
				db.Close()
			}

			// Check primary
			if primary, err := poolManager.GetPrimary(); err == nil {
				checkDB("primary", primary)
			} else {
				fmt.Printf("✗ primary: Failed to get connection - %v\n", err)
				unhealthyCount++
			}

			// Check replica (if different from primary)
			if replica, err := poolManager.GetReplica(); err == nil {
				checkDB("replica", replica)
			}

			fmt.Printf("\nSummary:\n")
			fmt.Printf("  Healthy:   %d\n", healthyCount)
			fmt.Printf("  Unhealthy: %d\n", unhealthyCount)
			fmt.Printf("  Total:     %d\n", healthyCount+unhealthyCount)

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
