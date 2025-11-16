// Copyright 2023 IAC. All Rights Reserved.

package commands

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func NewBackupCommand() *cobra.Command {
	var (
		dbType     string
		host       string
		port       int
		database   string
		username   string
		password   string
		outputFile string
		compress   bool
	)

	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup database",
		Long:  `Create a backup of the specified database`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if outputFile == "" {
				outputFile = fmt.Sprintf("%s_%s_backup.sql", database, time.Now().Format("20060102_150405"))
			}

			fmt.Printf("Creating backup of %s database...\n", dbType)
			fmt.Printf("  Host:       %s:%d\n", host, port)
			fmt.Printf("  Database:   %s\n", database)
			fmt.Printf("  Output:     %s\n", outputFile)
			fmt.Printf("  Compressed: %v\n", compress)

			fmt.Println("\nBackup commands by database type:")

			switch dbType {
			case "mysql":
				cmd := fmt.Sprintf("mysqldump -h %s -P %d -u %s -p%s %s > %s",
					host, port, username, password, database, outputFile)
				if compress {
					cmd += " | gzip"
					outputFile += ".gz"
				}
				fmt.Printf("\nMySQL:\n  %s\n", cmd)

			case "postgres":
				cmd := fmt.Sprintf("pg_dump -h %s -p %d -U %s -d %s -F c -f %s",
					host, port, username, database, outputFile)
				fmt.Printf("\nPostgreSQL:\n  PGPASSWORD=%s %s\n", password, cmd)

			case "mssql":
				fmt.Printf("\nMSSQL:\n")
				fmt.Printf("  BACKUP DATABASE [%s] TO DISK = '%s'\n", database, outputFile)

			case "oracle":
				fmt.Printf("\nOracle:\n")
				fmt.Printf("  expdp %s/%s@%s:%d/%s directory=BACKUP_DIR dumpfile=%s\n",
					username, password, host, port, database, outputFile)
			}

			fmt.Println("\nNote: This command shows the backup syntax.")
			fmt.Println("Actual backup execution will be implemented in future versions.")

			return nil
		},
	}

	cmd.Flags().StringVarP(&dbType, "type", "t", "mysql", "Database type")
	cmd.Flags().StringVarP(&host, "host", "H", "localhost", "Database host")
	cmd.Flags().IntVarP(&port, "port", "p", 3306, "Database port")
	cmd.Flags().StringVarP(&database, "database", "d", "", "Database name")
	cmd.Flags().StringVarP(&username, "username", "u", "", "Database username")
	cmd.Flags().StringVarP(&password, "password", "P", "", "Database password")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: auto-generated)")
	cmd.Flags().BoolVarP(&compress, "compress", "c", false, "Compress backup")

	cmd.MarkFlagRequired("database")
	cmd.MarkFlagRequired("username")

	return cmd
}

func NewRestoreCommand() *cobra.Command {
	var (
		dbType     string
		host       string
		port       int
		database   string
		username   string
		password   string
		inputFile  string
		compressed bool
	)

	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore database",
		Long:  `Restore a database from backup`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Restoring %s database from backup...\n", dbType)
			fmt.Printf("  Host:       %s:%d\n", host, port)
			fmt.Printf("  Database:   %s\n", database)
			fmt.Printf("  Input:      %s\n", inputFile)
			fmt.Printf("  Compressed: %v\n", compressed)

			fmt.Println("\nRestore commands by database type:")

			switch dbType {
			case "mysql":
				cmd := fmt.Sprintf("mysql -h %s -P %d -u %s -p%s %s < %s",
					host, port, username, password, database, inputFile)
				if compressed {
					cmd = fmt.Sprintf("gunzip < %s | %s", inputFile, cmd)
				}
				fmt.Printf("\nMySQL:\n  %s\n", cmd)

			case "postgres":
				cmd := fmt.Sprintf("pg_restore -h %s -p %d -U %s -d %s %s",
					host, port, username, database, inputFile)
				fmt.Printf("\nPostgreSQL:\n  PGPASSWORD=%s %s\n", password, cmd)

			case "mssql":
				fmt.Printf("\nMSSQL:\n")
				fmt.Printf("  RESTORE DATABASE [%s] FROM DISK = '%s'\n", database, inputFile)

			case "oracle":
				fmt.Printf("\nOracle:\n")
				fmt.Printf("  impdp %s/%s@%s:%d/%s directory=BACKUP_DIR dumpfile=%s\n",
					username, password, host, port, database, inputFile)
			}

			fmt.Println("\nNote: This command shows the restore syntax.")
			fmt.Println("Actual restore execution will be implemented in future versions.")
			fmt.Println("\nWARNING: Restoring will overwrite existing data!")

			return nil
		},
	}

	cmd.Flags().StringVarP(&dbType, "type", "t", "mysql", "Database type")
	cmd.Flags().StringVarP(&host, "host", "H", "localhost", "Database host")
	cmd.Flags().IntVarP(&port, "port", "p", 3306, "Database port")
	cmd.Flags().StringVarP(&database, "database", "d", "", "Database name")
	cmd.Flags().StringVarP(&username, "username", "u", "", "Database username")
	cmd.Flags().StringVarP(&password, "password", "P", "", "Database password")
	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input file")
	cmd.Flags().BoolVarP(&compressed, "compressed", "c", false, "Input is compressed")

	cmd.MarkFlagRequired("database")
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("input")

	return cmd
}
