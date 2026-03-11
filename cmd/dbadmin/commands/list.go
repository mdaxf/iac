// Copyright 2023 IAC. All Rights Reserved.

package commands

import (
	"fmt"

	dbconn "github.com/mdaxf/iac/databases"
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

			poolManager := dbconn.GetPoolManager()
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

			count := 0

			// List primary
			primary, err := poolManager.GetPrimary()
			if err == nil {
				count++
				fmt.Printf("%d. primary (%s)\n", count, primary.GetType())

				if verbose {
					status := "Healthy"
					if err := primary.Ping(); err != nil {
						status = "Unhealthy"
					}
					fmt.Printf("   Status:  %s\n", status)
					fmt.Printf("   Dialect: %s\n", primary.GetDialect())
					primary.Close()
					fmt.Println()
				}
			}

			// List replica
			replica, err := poolManager.GetReplica()
			if err == nil {
				count++
				fmt.Printf("%d. replica (%s)\n", count, replica.GetType())

				if verbose {
					status := "Healthy"
					if err := replica.Ping(); err != nil {
						status = "Unhealthy"
					}
					fmt.Printf("   Status:  %s\n", status)
					fmt.Printf("   Dialect: %s\n", replica.GetDialect())
					replica.Close()
					fmt.Println()
				}
			}

			if count == 0 {
				fmt.Println("No databases configured")
				return nil
			}

			fmt.Printf("\nConfigured Databases: %d\n", count)

			if !verbose {
				fmt.Println("\nUse --verbose to see detailed information")
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")

	return cmd
}
