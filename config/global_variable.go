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

package config

import (
	"github.com/mdaxf/iac/framework/cache"
)

var (
	SessionCache        cache.Cache
	SessionCacheTimeout int64
)

var GlobalConfiguration *GlobalConfig

type GlobalConfig struct {
	LogConfig          map[string]interface{}   `json:"log"`
	DocumentConfig     map[string]interface{}   `json:"documentdb"`
	DatabaseConfig     map[string]interface{}   `json:"database"`
	AltDatabasesConfig []map[string]interface{} `json:"altdatabases"`
	CacheConfig        map[string]interface{}   `json:"cache"`
	TranslationConfig  map[string]interface{}   `json:"translation"`
}
