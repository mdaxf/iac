// Copyright 2023 IAC. All Rights Reserved.

package commands

import (
	"fmt"

	"github.com/mdaxf/iac/dbinitializer"
	"github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured databases",
		Long:  `List all databases configured in the environment`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize databases from environment
			dbInit := dbinitializer.NewDatabaseInitializer()
			if err := dbInit.InitializeFromEnvironment(); err != nil {
				return fmt.Errorf("failed to initialize databases: %w", err)
			}

			poolManager := dbInit.GetPoolManager()
			if poolManager == nil {
				fmt.Println("No databases configured")
				fmt.Println("\nTo configure databases, set environment variables:")
				fmt.Println("  DB_TYPE=mysql")
				fmt.Println("  DB_HOST=localhost")
				fmt.Println("  DB_PORT=3306")
				fmt.Println("  DB_DATABASE=iac")
				fmt.Println("  DB_USERNAME=iac_user")
				fmt.Println("  DB_PASSWORD=iac_pass")
				return nil
			}

			databases := poolManager.GetAllDatabases()
			if len(databases) == 0 {
				fmt.Println("No databases configured")
				return nil
			}

			fmt.Printf("Configured Databases (%d):\n\n", len(databases))

			for i, dbType := range databases {
				fmt.Printf("%d. %s\n", i+1, dbType)

				if verbose {
					db, err := poolManager.GetPrimary(dbType)
					if err != nil {
						fmt.Printf("   Status:  Error - %v\n", err)
					} else {
						status := "Healthy"
						if err := db.Ping(); err != nil {
							status = "Unhealthy"
						}

						fmt.Printf("   Status:  %s\n", status)
						fmt.Printf("   Dialect: %s\n", db.GetDialect())

						db.Close()
					}

					// Check for replicas
					replicas, err := poolManager.GetReplicas(dbType)
					if err == nil && len(replicas) > 0 {
						fmt.Printf("   Replicas: %d\n", len(replicas))
					}

					fmt.Println()
				}
			}

			if !verbose {
				fmt.Println("\nUse --verbose to see detailed information")
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")

	return cmd
}
