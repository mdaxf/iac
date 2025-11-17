// Copyright 2023 IAC. All Rights Reserved.

package commands

import (
	"context"
	"fmt"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
	"github.com/spf13/cobra"
)

func NewConnectCommand() *cobra.Command {
	var (
		dbType   string
		host     string
		port     int
		database string
		username string
		password string
		sslMode  string
	)

	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Test database connection",
		Long:  `Test connection to a database and verify credentials`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := dbconn.DBConfig{
				Type:         dbType,
				Host:         host,
				Port:         port,
				Database:     database,
				Username:     username,
				Password:     password,
				MaxIdleConns: 5,
				MaxOpenConns: 10,
				ConnTimeout:  30,
				Options:      make(map[string]string),
			}

			if sslMode != "" {
				config.Options["sslmode"] = sslMode
			}

			fmt.Printf("Connecting to %s database at %s:%d...\n", dbType, host, port)

			db, err := dbconn.NewRelationalDB(config)
			if err != nil {
				return fmt.Errorf("failed to create database instance: %w", err)
			}
			defer db.Close()

			start := time.Now()
			if err := db.Connect(config); err != nil {
				return fmt.Errorf("connection failed: %w", err)
			}

			duration := time.Since(start)
			fmt.Printf("✓ Connected successfully in %v\n", duration)

			// Test ping
			start = time.Now()
			if err := db.Ping(); err != nil {
				return fmt.Errorf("ping failed: %w", err)
			}

			pingDuration := time.Since(start)
			fmt.Printf("✓ Ping successful in %v\n", pingDuration)

			// Get dialect
			dialect := db.GetDialect()
			fmt.Printf("\nDatabase Information:\n")
			fmt.Printf("  Type:     %s\n", dbType)
			fmt.Printf("  Dialect:  %s\n", dialect)
			fmt.Printf("  Host:     %s:%d\n", host, port)
			fmt.Printf("  Database: %s\n", database)

			// Test features
			fmt.Printf("\nSupported Features:\n")
			features := []string{"transactions", "jsonb", "cte", "window_functions", "fulltext", "arrays"}
			for _, feature := range features {
				if db.SupportsFeature(feature) {
					fmt.Printf("  ✓ %s\n", feature)
				} else {
					fmt.Printf("  ✗ %s\n", feature)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&dbType, "type", "t", "mysql", "Database type (mysql, postgres, mssql, oracle)")
	cmd.Flags().StringVarP(&host, "host", "H", "localhost", "Database host")
	cmd.Flags().IntVarP(&port, "port", "p", 3306, "Database port")
	cmd.Flags().StringVarP(&database, "database", "d", "", "Database name")
	cmd.Flags().StringVarP(&username, "username", "u", "", "Database username")
	cmd.Flags().StringVarP(&password, "password", "P", "", "Database password")
	cmd.Flags().StringVar(&sslMode, "ssl-mode", "", "SSL mode (disable, require, verify-ca, verify-full)")

	cmd.MarkFlagRequired("database")
	cmd.MarkFlagRequired("username")

	return cmd
}
