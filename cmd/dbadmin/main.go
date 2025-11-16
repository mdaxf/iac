// Copyright 2023 IAC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"

	"github.com/mdaxf/iac/cmd/dbadmin/commands"
	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"
	cfgFile string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "dbadmin",
		Short: "IAC Database Administration Tool",
		Long: `IAC Database Administration Tool

A comprehensive CLI tool for managing IAC databases including:
- Connection testing and health checks
- Migration execution
- Backup and restore operations
- Performance monitoring
- Schema management`,
		Version: version,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.iacdb.yaml)")

	// Add commands
	rootCmd.AddCommand(commands.NewConnectCommand())
	rootCmd.AddCommand(commands.NewHealthCommand())
	rootCmd.AddCommand(commands.NewMigrateCommand())
	rootCmd.AddCommand(commands.NewBackupCommand())
	rootCmd.AddCommand(commands.NewRestoreCommand())
	rootCmd.AddCommand(commands.NewSchemaCommand())
	rootCmd.AddCommand(commands.NewMetricsCommand())
	rootCmd.AddCommand(commands.NewListCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
