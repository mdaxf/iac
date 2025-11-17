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
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	dbconn "github.com/mdaxf/iac/databases"
	deploymgr "github.com/mdaxf/iac/deployment/deploy"
	"github.com/mdaxf/iac/deployment/devops"
	"github.com/mdaxf/iac/deployment/models"
	packagemgr "github.com/mdaxf/iac/deployment/package"
	"github.com/mdaxf/iac/documents"
)

func main() {
	// Define commands
	packageCmd := flag.NewFlagSet("package", flag.ExitOnError)
	deployCmd := flag.NewFlagSet("deploy", flag.ExitOnError)
	versionCmd := flag.NewFlagSet("version", flag.ExitOnError)

	// Package flags
	packageType := packageCmd.String("type", "database", "Package type: database or document")
	packageName := packageCmd.String("name", "", "Package name")
	packageVersion := packageCmd.String("version", "1.0.0", "Package version")
	packageTables := packageCmd.String("tables", "", "Comma-separated table names")
	packageCollections := packageCmd.String("collections", "", "Comma-separated collection names")
	packageOutput := packageCmd.String("output", "package.json", "Output file")

	// Deploy flags
	deployType := deployCmd.String("type", "database", "Deploy type: database or document")
	deployInput := deployCmd.String("input", "", "Package file to deploy")
	deployUpdate := deployCmd.Bool("update", false, "Update existing records")
	deployDryRun := deployCmd.Bool("dry-run", false, "Dry run mode")

	// Version control flags
	versionAction := versionCmd.String("action", "commit", "Action: commit, branch, merge, tag, list")
	versionObjectType := versionCmd.String("object-type", "TranCode", "Object type")
	versionObjectID := versionCmd.String("object-id", "", "Object ID")
	versionBranch := versionCmd.String("branch", "main", "Branch name")
	versionMessage := versionCmd.String("message", "", "Commit message")

	if len(os.Args) < 2 {
		fmt.Println("Usage: deploy_cli [command] [options]")
		fmt.Println("\nCommands:")
		fmt.Println("  package    Package database or document data")
		fmt.Println("  deploy     Deploy a package")
		fmt.Println("  version    Version control operations")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "package":
		packageCmd.Parse(os.Args[2:])
		handlePackage(*packageType, *packageName, *packageVersion, *packageTables, *packageCollections, *packageOutput)

	case "deploy":
		deployCmd.Parse(os.Args[2:])
		handleDeploy(*deployType, *deployInput, *deployUpdate, *deployDryRun)

	case "version":
		versionCmd.Parse(os.Args[2:])
		handleVersion(*versionAction, *versionObjectType, *versionObjectID, *versionBranch, *versionMessage)

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func handlePackage(pkgType, name, version, tables, collections, output string) {
	fmt.Printf("Packaging %s: %s v%s\n", pkgType, name, version)

	user := "cli-user"

	if pkgType == "database" {
		// Database packaging
		if tables == "" {
			fmt.Println("Error: --tables is required for database packaging")
			os.Exit(1)
		}

		// Initialize database
		dbconn.InitializeDB()

		dbTx, err := dbconn.DB.Begin()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer dbTx.Rollback()

		packager := packagemgr.NewDatabasePackager(user, dbTx, dbconn.DatabaseType)

		filter := models.PackageFilter{
			Tables: splitString(tables),
		}

		pkg, err := packager.PackageTables(name, version, user, filter)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		data, err := packager.ExportPackage(pkg)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if err := ioutil.WriteFile(output, data, 0644); err != nil {
			fmt.Printf("Error writing file: %v\n", err)
			os.Exit(1)
		}

		dbTx.Commit()

		fmt.Printf("Package created: %s (%d tables, %d KB)\n", output, len(pkg.DatabaseData.Tables), len(data)/1024)

	} else if pkgType == "document" {
		// Document packaging
		if collections == "" {
			fmt.Println("Error: --collections is required for document packaging")
			os.Exit(1)
		}

		// Initialize document database
		docDB, err := documents.InitMongoDB(documents.DatabaseConnection, documents.DatabaseName)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		packager := packagemgr.NewDocumentPackager(docDB, user)

		filter := models.PackageFilter{
			Collections: splitString(collections),
		}

		pkg, err := packager.PackageCollections(name, version, user, filter)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		data, err := packager.ExportPackage(pkg)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if err := ioutil.WriteFile(output, data, 0644); err != nil {
			fmt.Printf("Error writing file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Package created: %s (%d collections, %d KB)\n", output, len(pkg.DocumentData.Collections), len(data)/1024)
	}
}

func handleDeploy(deployType, input string, update, dryRun bool) {
	fmt.Printf("Deploying %s package: %s\n", deployType, input)

	if input == "" {
		fmt.Println("Error: --input is required")
		os.Exit(1)
	}

	data, err := ioutil.ReadFile(input)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	user := "cli-user"

	options := models.DeploymentOptions{
		UpdateExisting:  update,
		DryRun:          dryRun,
		BatchSize:       100,
		ContinueOnError: false,
	}

	if deployType == "database" {
		// Database deployment
		dbconn.InitializeDB()

		dbTx, err := dbconn.DB.Begin()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer dbTx.Rollback()

		packager := packagemgr.NewDatabasePackager(user, nil, dbconn.DatabaseType)
		pkg, err := packager.ImportPackage(data)
		if err != nil {
			fmt.Printf("Error importing package: %v\n", err)
			os.Exit(1)
		}

		deployer := deploymgr.NewDatabaseDeployer(user, dbTx, dbconn.DatabaseType)
		record, err := deployer.Deploy(pkg, options)
		if err != nil {
			fmt.Printf("Error deploying: %v\n", err)
			os.Exit(1)
		}

		if record.Status == "completed" {
			dbTx.Commit()
			fmt.Printf("Deployment successful: %s\n", record.ID)
			fmt.Printf("Tables deployed: %d\n", len(pkg.DatabaseData.Tables))
		} else {
			fmt.Printf("Deployment status: %s\n", record.Status)
			if len(record.ErrorLog) > 0 {
				fmt.Println("Errors:")
				for _, err := range record.ErrorLog {
					fmt.Printf("  - %s\n", err)
				}
			}
			os.Exit(1)
		}

	} else if deployType == "document" {
		// Document deployment
		docDB, err := documents.InitMongoDB(documents.DatabaseConnection, documents.DatabaseName)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		packager := packagemgr.NewDocumentPackager(docDB, user)
		pkg, err := packager.ImportPackage(data)
		if err != nil {
			fmt.Printf("Error importing package: %v\n", err)
			os.Exit(1)
		}

		deployer := deploymgr.NewDocumentDeployer(docDB, user)
		record, err := deployer.Deploy(pkg, options)
		if err != nil {
			fmt.Printf("Error deploying: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Deployment %s: %s\n", record.Status, record.ID)
		fmt.Printf("Collections deployed: %d\n", len(pkg.DocumentData.Collections))
	}
}

func handleVersion(action, objectType, objectID, branch, message string) {
	fmt.Printf("Version control: %s\n", action)

	docDB, err := documents.InitMongoDB(documents.DatabaseConnection, documents.DatabaseName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	user := "cli-user"
	vc := devops.NewVersionControl(docDB, user)

	switch action {
	case "commit":
		if objectID == "" {
			fmt.Println("Error: --object-id is required")
			os.Exit(1)
		}
		if message == "" {
			fmt.Println("Error: --message is required")
			os.Exit(1)
		}

		// Get current document
		collection := docDB.MongoDBDatabase.Collection(objectType)
		var doc map[string]interface{}
		// Simplified - would need proper ObjectID parsing
		// err := collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&doc)

		fmt.Printf("Would commit %s: %s to branch %s\n", objectType, objectID, branch)
		fmt.Printf("Message: %s\n", message)

	case "list":
		if objectID == "" {
			fmt.Println("Error: --object-id is required")
			os.Exit(1)
		}

		versions, err := vc.ListVersions(devops.ObjectType(objectType), objectID, branch)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Versions for %s %s:\n", objectType, objectID)
		for _, v := range versions {
			fmt.Printf("  %s - %s (%s) - %s\n", v.Version, v.CommitMessage, v.CreatedBy, v.CreatedAt)
		}

	default:
		fmt.Printf("Unknown action: %s\n", action)
		os.Exit(1)
	}
}

func splitString(s string) []string {
	if s == "" {
		return []string{}
	}
	result := []string{}
	for _, item := range splitByComma(s) {
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func splitByComma(s string) []string {
	result := []string{}
	current := ""
	for _, c := range s {
		if c == ',' {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
