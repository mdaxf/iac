// Copyright 2023 IAC. All Rights Reserved.

package commands

import (
	"context"
	"fmt"

	"github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/services"
	"github.com/spf13/cobra"
)

func NewSchemaCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Schema operations",
		Long:  `Perform schema discovery and management operations`,
	}

	cmd.AddCommand(newSchemaDiscoverCommand())
	cmd.AddCommand(newSchemaListCommand())

	return cmd
}

func newSchemaDiscoverCommand() *cobra.Command{
	var (
		dbType     string
		host       string
		port       int
		database   string
		schema     string
		username   string
		password   string
		sslMode    string
	)

	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover database schema",
		Long:  `Discover all tables and columns in a database`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := databases.DBConfig{
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

			fmt.Printf("Discovering schema for %s database...\n", dbType)

			db, err := databases.NewRelationalDB(config)
			if err != nil {
				return fmt.Errorf("failed to create database: %w", err)
			}
			defer db.Close()

			if err := db.Connect(config); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}

			dialect := db.GetDialect()
			if schema == "" {
				schema = database
			}

			// Get tables
			tablesQuery := services.GetTablesQuery(dialect, schema)
			rows, err := db.Query(tablesQuery)
			if err != nil {
				return fmt.Errorf("failed to query tables: %w", err)
			}
			defer rows.Close()

			fmt.Printf("\nTables in schema '%s':\n\n", schema)

			tableCount := 0
			for rows.Next() {
				var tableName, tableComment, tableSchema string
				if err := rows.Scan(&tableName, &tableComment, &tableSchema); err != nil {
					continue
				}

				tableCount++
				fmt.Printf("%d. %s", tableCount, tableName)
				if tableComment != "" {
					fmt.Printf(" - %s", tableComment)
				}
				fmt.Println()

				// Get columns for this table
				columnsQuery := services.GetColumnsQuery(dialect, schema, tableName)
				colRows, err := db.Query(columnsQuery)
				if err != nil {
					fmt.Printf("   Error getting columns: %v\n", err)
					continue
				}

				colCount := 0
				for colRows.Next() {
					var colName, dataType, isNullable, colKey, colComment string
					var ordinalPos int
					if err := colRows.Scan(&colName, &dataType, &isNullable, &colKey, &colComment, &ordinalPos); err != nil {
						continue
					}

					colCount++
					nullable := ""
					if isNullable == "YES" {
						nullable = "NULL"
					} else {
						nullable = "NOT NULL"
					}

					keyInfo := ""
					if colKey == "PRI" {
						keyInfo = " [PRIMARY KEY]"
					}

					fmt.Printf("   - %s (%s) %s%s", colName, dataType, nullable, keyInfo)
					if colComment != "" {
						fmt.Printf(" - %s", colComment)
					}
					fmt.Println()
				}
				colRows.Close()

				if colCount == 0 {
					fmt.Println("   No columns found")
				}
				fmt.Println()
			}

			if tableCount == 0 {
				fmt.Println("No tables found")
			} else {
				fmt.Printf("Total tables: %d\n", tableCount)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&dbType, "type", "t", "mysql", "Database type")
	cmd.Flags().StringVarP(&host, "host", "H", "localhost", "Database host")
	cmd.Flags().IntVarP(&port, "port", "p", 3306, "Database port")
	cmd.Flags().StringVarP(&database, "database", "d", "", "Database name")
	cmd.Flags().StringVarP(&schema, "schema", "s", "", "Schema name (default: same as database)")
	cmd.Flags().StringVarP(&username, "username", "u", "", "Database username")
	cmd.Flags().StringVarP(&password, "password", "P", "", "Database password")
	cmd.Flags().StringVar(&sslMode, "ssl-mode", "", "SSL mode")

	cmd.MarkFlagRequired("database")
	cmd.MarkFlagRequired("username")

	return cmd
}

func newSchemaListCommand() *cobra.Command {
	var (
		dbType   string
		host     string
		port     int
		username string
		password string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all databases/schemas",
		Long:  `List all available databases or schemas`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := databases.DBConfig{
				Type:         dbType,
				Host:         host,
				Port:         port,
				Database:     "",
				Username:     username,
				Password:     password,
				MaxIdleConns: 5,
				MaxOpenConns: 10,
				ConnTimeout:  30,
				Options:      make(map[string]string),
			}

			db, err := databases.NewRelationalDB(config)
			if err != nil {
				return fmt.Errorf("failed to create database: %w", err)
			}
			defer db.Close()

			// Connect without specific database
			if dbType == "mysql" {
				config.Database = "information_schema"
			} else if dbType == "postgres" {
				config.Database = "postgres"
			} else {
				config.Database = "master"
			}

			if err := db.Connect(config); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}

			dialect := db.GetDialect()
			query := services.GetDatabaseListQuery(dialect)

			rows, err := db.Query(query)
			if err != nil {
				return fmt.Errorf("failed to list databases: %w", err)
			}
			defer rows.Close()

			fmt.Printf("Databases on %s:\n\n", host)

			count := 0
			for rows.Next() {
				var dbName string
				if err := rows.Scan(&dbName); err != nil {
					continue
				}

				count++
				fmt.Printf("%d. %s\n", count, dbName)
			}

			if count == 0 {
				fmt.Println("No databases found")
			} else {
				fmt.Printf("\nTotal: %d databases\n", count)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&dbType, "type", "t", "mysql", "Database type")
	cmd.Flags().StringVarP(&host, "host", "H", "localhost", "Database host")
	cmd.Flags().IntVarP(&port, "port", "p", 3306, "Database port")
	cmd.Flags().StringVarP(&username, "username", "u", "", "Database username")
	cmd.Flags().StringVarP(&password, "password", "P", "", "Database password")

	cmd.MarkFlagRequired("username")

	return cmd
}
