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

package handlers

import (
	"database/sql"

	"github.com/mdaxf/iac/documents"
)

// HandlerFunc defines the signature for job handler functions
type HandlerFunc func(inputs map[string]interface{}, tx *sql.Tx, docDB *documents.DocDB) (map[string]interface{}, error)

// HandlerRegistry maps handler names to their implementation functions
var HandlerRegistry = map[string]HandlerFunc{
	"PACKAGE_DEPLOYMENT": PackageDeploymentHandler,
	"PACKAGE_GENERATION": PackageGenerationHandler,
}

// GetHandler retrieves a handler function by name
func GetHandler(name string) (HandlerFunc, bool) {
	handler, exists := HandlerRegistry[name]
	return handler, exists
}

// RegisterHandler registers a new handler function
func RegisterHandler(name string, handler HandlerFunc) {
	HandlerRegistry[name] = handler
}

// ListHandlers returns all registered handler names
func ListHandlers() []string {
	handlers := make([]string, 0, len(HandlerRegistry))
	for name := range HandlerRegistry {
		handlers = append(handlers, name)
	}
	return handlers
}
