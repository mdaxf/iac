# IAC Codebase - Comprehensive Database Architecture Overview

**Date:** November 17, 2025
**Project:** IAC (Integrated Application Components) Framework
**Status:** Multi-Database Support Implementation in Progress

---

## Executive Summary

The IAC framework is a comprehensive Go-based application platform with sophisticated multi-layer database architecture supporting:
- **Relational Databases:** MySQL, PostgreSQL, MSSQL, Oracle
- **Document Databases:** MongoDB (primary), PostgreSQL JSONB (emerging support)
- **Connection Pooling & Management:** Multi-database connection management with failover support
- **Schema Versioning:** Comprehensive migration system with checksum validation
- **Deployment Mechanism:** Docker-based deployment with automated database initialization

---

## 1. DATABASE LAYERS AND ARCHITECTURE

### 1.1 Relational Database Layer

**Location:** `/home/user/iac/databases/`

#### Core Components:

**Interface-Based Abstraction (`interface.go`)**
```go
type RelationalDB interface {
    Connect(config *DBConfig) error
    Close() error
    Ping() error
    DB() *sql.DB
    
    // Query Operations
    Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
    QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
    Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
    
    // Database Information & Schema Operations
    GetDialect() Dialect
    GetType() DBType
    SupportsFeature(feature Feature) bool
    ListTables(ctx context.Context, schema string) ([]string, error)
    TableExists(ctx context.Context, schema, tableName string) (bool, error)
    GetTableSchema(ctx context.Context, schema, tableName string) (*TableSchema, error)
}
```

**Supported Database Types:**
- `DBTypeMySQL` - MySQL (default)
- `DBTypePostgreSQL` - PostgreSQL
- `DBTypeMSSQL` - Microsoft SQL Server
- `DBTypeOracle` - Oracle Database

#### Configuration Structure (`DBConfig`)
```go
type DBConfig struct {
    Type              DBType              // mysql, postgres, mssql, oracle
    Host              string              // Database server hostname
    Port              int                 // Connection port
    Database          string              // Database name
    Schema            string              // Optional schema name
    Username          string              // Authentication username
    Password          string              // Authentication password
    
    // SSL/TLS Configuration
    SSLMode           string              // SSL mode (disable, allow, prefer, require)
    SSLCert           string              // Client certificate path
    SSLKey            string              // Client key path
    SSLRootCert       string              // Root CA certificate path
    
    // Connection Pooling
    MaxIdleConns      int                 // Default: 5
    MaxOpenConns      int                 // Default: 10
    ConnMaxLifetime   time.Duration       // Connection lifetime
    ConnMaxIdleTime   time.Duration       // Idle timeout
    ConnTimeout       int                 // Connection timeout in seconds
    
    // Database-Specific Options
    Options           map[string]string   // Driver-specific parameters
}
```

#### Database Factory Pattern (`factory.go`)

The factory pattern provides database instantiation with automatic driver registration:

```go
type DatabaseFactory struct {
    drivers         map[DBType]DriverConstructor    // Driver constructors
    dialects        map[DBType]Dialect               // SQL dialects
    mu              sync.RWMutex                     // Thread safety
    instances       map[string]RelationalDB          // Connection cache
}

// Usage
factory := GetFactory()
db, err := factory.NewRelationalDB(config)
```

**Key Features:**
- Singleton pattern for factory instance
- Thread-safe driver registration
- Connection caching and reuse
- Automatic dialect selection

#### SQL Dialect System (`dialects.go`)

The dialect interface handles database-specific SQL generation:

```go
type Dialect interface {
    // Identifier quoting
    QuoteIdentifier(name string) string
    
    // Parameter placeholders
    Placeholder(n int) string
    
    // Pagination syntax
    LimitOffset(limit, offset int) string
    
    // Type mapping
    DataTypeMapping(genericType string) string
    ConvertValue(value interface{}, targetType string) (interface{}, error)
    
    // Feature detection
    SupportsReturning() bool
    SupportsUpsert() bool
    SupportsCTE() bool
    SupportsJSON() bool
    SupportsFullTextSearch() bool
    
    // DDL Generation
    CreateTableDDL(schema *TableSchema) string
    AddColumnDDL(tableName string, column *ColumnInfo) string
    CreateIndexDDL(tableName string, index *IndexInfo) string
}
```

**Dialect-Specific Implementations:**
- **MySQL:** `GROUP_CONCAT`, `LIMIT OFFSET` pagination
- **PostgreSQL:** `ILIKE` case-insensitive, `RETURNING` clause, `JSONB` support
- **MSSQL:** `OFFSET ROWS FETCH` pagination, `GETDATE()` function
- **Oracle:** `FETCH FIRST N ROWS ONLY`, `SYSTIMESTAMP`, schema-aware

#### Connection Pool Management (`poolmanager.go`)

Multi-database connection pool management:

```go
type PoolManager struct {
    pools            map[string]*ConnectionPool
    selector         *DatabaseSelector
    primary          RelationalDB
    replicas         []RelationalDB
}

Features:
- Per-database connection pooling
- Read/write splitting
- Replica load balancing
- Connection lifecycle management
- Pool statistics and monitoring
```

#### Database Operations Layer (`dboperation.go`)

High-level database operations:

```go
type DBOperation struct {
    DBTx           *sql.Tx                // Current transaction
    ModuleName     string                 // Module name for logging
    iLog           logger.Log             // Logger instance
    User           string                 // User performing operation
}

Methods:
- Query()          // Execute SELECT queries
- Exec()           // Execute INSERT/UPDATE/DELETE
- QueryRow()       // Single row queries
- BeginTx()        // Start transactions
- GetDialect()     // Get SQL dialect for current DB
- QuoteIdentifier() // Quote database identifiers
```

### 1.2 Document Database Layer

**Location:** `/home/user/iac/documents/`

#### MongoDB Integration (`mongodb.go`)

Primary document database implementation:

```go
type DocDB struct {
    MongoDBClient       *mongo.Client
    MongoDBDatabase     *mongo.Database
    MongoDBCollection_TC *mongo.Collection
    
    DatabaseType        string              // "mongodb"
    DatabaseConnection  string              // Connection string (MongoDB URI)
    DatabaseName        string              // Database name
    iLog                logger.Log          // Logger
    monitoring          bool                // Health check status
}

Initialization:
func InitMongoDB(conn string, dbName string) (*DocDB, error)
    - Creates MongoDB client with connection string
    - Validates connection with Ping()
    - Starts monitoring goroutine for reconnection
```

#### Document Database Interface (`interface.go`)

Abstraction for document database operations:

```go
type DocumentDB interface {
    // Connection Management
    Connect(config *DocDBConfig) error
    Close() error
    Ping(ctx context.Context) error
    
    // Collection Operations
    CreateCollection(ctx context.Context, name string) error
    DropCollection(ctx context.Context, name string) error
    ListCollections(ctx context.Context) ([]string, error)
    CollectionExists(ctx context.Context, name string) (bool, error)
    
    // Document CRUD
    InsertOne(ctx context.Context, collection string, document interface{}) (string, error)
    InsertMany(ctx context.Context, collection string, documents []interface{}) ([]string, error)
    FindOne(ctx context.Context, collection string, filter interface{}) (map[string]interface{}, error)
    FindMany(ctx context.Context, collection string, filter interface{}, opts *FindOptions) ([]map[string]interface{}, error)
    UpdateOne(ctx context.Context, collection string, filter interface{}, update interface{}) error
    UpdateMany(ctx context.Context, collection string, filter interface{}, update interface{}) (int64, error)
    DeleteOne(ctx context.Context, collection string, filter interface{}) error
    DeleteMany(ctx context.Context, collection string, filter interface{}) (int64, error)
    
    // Index & Aggregation
    CreateIndex(ctx context.Context, collection string, keys map[string]int, options *IndexOptions) error
    Aggregate(ctx context.Context, collection string, pipeline []map[string]interface{}) ([]map[string]interface{}, error)
}
```

#### Configuration (`DocDBConfig`)

```go
type DocDBConfig struct {
    Type            DocDBType               // mongodb, postgres
    Host            string                  // Server hostname
    Port            int                     // Connection port (27017 for MongoDB)
    Database        string                  // Database name
    Username        string                  // Authentication username
    Password        string                  // Authentication password
    
    SSLMode         string                  // SSL mode for MongoDB
    AuthSource      string                  // MongoDB auth database (default: admin)
    ReplicaSet      string                  // MongoDB replica set name
    
    MaxPoolSize     int                     // Default: 100
    MinPoolSize     int                     // Default: 10
    ConnTimeout     int                     // Timeout in seconds
    
    Options         map[string]string       // Connection-specific options
}
```

#### Core MongoDB Operations

**Query Collection:**
```go
func (doc *DocDB) QueryCollection(collectionname string, filter bson.M, projection bson.M) ([]bson.M, error)
    - Queries documents with optional filter and projection
    - Returns array of BSON documents
    - Used for complex queries with multiple results
```

**Get by Identifier:**
```go
func (doc *DocDB) GetItembyID(collectionname string, id string) (bson.M, error)
    - Retrieves single document by MongoDB ObjectID
    - Converts hex string to ObjectID

func (doc *DocDB) GetItembyUUID(collectionname string, uuid string) (bson.M, error)
    - Retrieves document by custom UUID field
    
func (doc *DocDB) GetDefaultItembyName(collectionname string, name string) (bson.M, error)
    - Retrieves default version of named item
```

**Insert Operations:**
```go
func (doc *DocDB) InsertCollection(collectionname string, idata interface{}) (*mongo.InsertOneResult, error)
    - Inserts single document
    - Returns insertion result with generated ObjectID
```

**Update Operations:**
```go
func (doc *DocDB) UpdateCollection(collectionname string, filter bson.M, update bson.M, idata interface{}) error
    - Updates documents matching filter
    - Supports both UpdateOne and ReplaceOne operations
```

**Delete Operations:**
```go
func (doc *DocDB) DeleteItemFromCollection(collectionname string, documentid string) error
    - Deletes single document by ObjectID
```

#### Health Monitoring

```go
func (doc *DocDB) MonitorAndReconnect()
    - Runs as background goroutine
    - Pings MongoDB every 1 second
    - Automatically reconnects on failure
    - Implements exponential backoff (5 seconds retry delay)
```

---

## 2. DOCUMENT DATABASE USAGE AND COLLECTIONS

### 2.1 Primary MongoDB Collections

**Main Collections Used:**

1. **Transaction_Code** - Core business transaction definitions
   - Stores TranCode definitions with execution logic
   - Fields: trancodename, version, isdefault, functiongroups, inputs, outputs
   - Contains function groups, routing rules, and test data
   - Filtering: `{trancodename: "...", isdefault: true}`

2. **WorkFlow** - Workflow definitions and orchestration
   - Stores workflow models with nodes and links
   - Fields: name, uuid, version, nodes, links
   - Node types: decision, activity, subprocess
   - Contains routing tables for conditional logic

3. **Process_Plan** - Process planning and execution
   - Stores process definitions
   - Field structure: name, description, steps
   - Integration with workflow execution

4. **Job_History** - Asynchronous job tracking
   - Records job execution history
   - Fields: job_id, status, created_at, completed_at
   - Used with framework queue system

5. **Notifications** - User notifications and alerts
   - Stores system notifications
   - Fields: recipient, type, message, read_status
   - Real-time notification delivery support

6. **3D_Models** (Models3D) - 3D model storage and metadata
   - Stores CAD/3D model references
   - Fields: name, type, geometry, metadata
   - Integration with visualization components

### 2.2 Database Configuration

**Default Configuration (configuration.json):**
```json
{
    "documentdb": {
        "type": "mongodb",
        "connection": "mongodb://localhost:27017",
        "database": "IAC_CFG"
    }
}
```

**Environment Variable Overrides:**
- `DOCDB_HOST` - MongoDB server hostname
- `DOCDB_USER` - Authentication username
- `DOCDB_PASSWORD` - Authentication password
- `DOCDB_NAME` - Database name

---

## 3. KEY DATA MODELS AND RELATIONSHIPS

### 3.1 Transaction Code (TranCode) Model

**Location:** `/home/user/iac/engine/types/types.go`

```go
type TranCode struct {
    ID              string                  // MongoDB ObjectID (_id)
    UUID            string                  // Unique identifier
    Name            string                  // trancodename field
    Version         string                  // Version string
    IsDefault       bool                    // isdefault flag
    Status          Status                  // Design/Test/Prototype/Production
    
    Inputs          []Input                 // Input parameters
    Outputs         []Output                // Output parameters
    Functiongroups  []FuncGroup             // Execution function groups
    Workflow        map[string]interface{}  // Associated workflow data
    Firstfuncgroup  string                  // Entry point function group
    
    SystemData      SystemData              // Audit/tracking data
    Description     string                  // Human-readable description
    TestDatas       []TestData              // Unit test definitions
}

type Input struct {
    ID              string                  // Input identifier
    Name            string                  // Parameter name
    Source          InputSource             // Constant/Function/Session
    Datatype        DataType                // String/Integer/Float/Bool/DateTime/Object
    Inivalue        string                  // Initial value
    Defaultvalue    string                  // Default value
    Value           string                  // Current value
    List            bool                    // Is array/list
    Repeat          bool                    // Can repeat
    Aliasname       string                  // Alias for reference
    Description     string                  // Parameter description
}

type Output struct {
    ID              string                  // Output identifier
    Name            string                  // Parameter name
    Outputdest      []OutputDest            // Destinations (Session/External)
    Datatype        DataType                // Output data type
    Inivalue        string                  // Initial value
    Defaultvalue    string                  // Default value
    Value           string                  // Current value
    List            bool                    // Is array/list
    Aliasname       []string                // Aliases
    Description     string                  // Parameter description
}
```

### 3.2 Function Group Model

```go
type FuncGroup struct {
    ID                  string                  // Unique identifier
    Name                string                  // Function group name
    Functions           []Function              // Functions in group
    Executionsequence   string                  // Serial/Parallel execution
    Session             map[string]interface{}  // Session variables
    RouterDef           RouterDef               // Routing rules
    functiongroupname   string                  // Name (duplicate field)
    Description         string                  // Group description
    routing             bool                    // Has routing logic
    Type                string                  // Type of function group
    x, y, width, height int                     // UI positioning
}

type Function struct {
    ID              string                  // Function identifier
    Name            string                  // Function name
    Version         string                  // Function version
    Status          Status                  // Status (Design/Test/Prototype/Production)
    Functype        FunctionType            // Type of function
    
    Inputs          []Input                 // Function inputs
    Outputs         []Output                // Function outputs
    
    Content         string                  // Function logic/SQL
    Script          string                  // JavaScript/GoExpr script
    Mapdata         map[string]interface{}  // Data mapping configuration
    FunctionName    string                  // Executable function name
    
    Description     string                  // Function description
    Type            string                  // Function type string
    x, y, width, height int                 // UI positioning
}
```

**Function Types:**
- `InputMap` - Input mapping
- `GoExpr` - Go expression evaluation
- `Javascript` - JavaScript execution
- `Query` - SQL query execution
- `StoreProcedure` - Stored procedure call
- `SubTranCode` - Sub-transaction code invocation
- `TableInsert/Update/Delete` - Table operations
- `CollectionInsert/Update/Delete` - MongoDB collection operations
- `ThrowError` - Error handling
- `SendMessage` - Message bus operations
- `SendEmail` - Email notifications
- `ExplodeWorkFlow` - Workflow integration
- `WebServiceCall` - REST API calls

### 3.3 Workflow Model

**Location:** `/home/user/iac/workflow/types/type.go`

```go
type WorkFlow struct {
    ID              primitive.ObjectID      // MongoDB ID
    Name            string                  // Workflow name
    UUID            string                  // Unique identifier
    Version         string                  // Workflow version
    Description     string                  // Workflow description
    ISDefault       bool                    // Is default version
    Type            string                  // Workflow type
    
    Nodes           []Node                  // Workflow nodes
    Links           []Link                  // Node connections
}

type Node struct {
    Name            string                  // Node name
    ID              string                  // Node identifier
    Description     string                  // Node description
    Type            string                  // Node type (activity/decision/etc)
    Page            string                  // Associated UI page
    TranCode        string                  // Associated transaction code
    
    Roles           []string                // Allowed roles
    Users           []string                // Allowed users
    Roleids         []int64                 // Role IDs
    Userids         []int64                 // User IDs
    
    PreCondition    map[string]interface{}  // Execution preconditions
    PostCondition   map[string]interface{}  // Execution postconditions
    ProcessData     map[string]interface{}  // Process data mapping
    
    RoutingTables   []RoutingTable          // Conditional routing rules
}

type Link struct {
    Name            string                  // Link name
    ID              string                  // Link identifier
    Type            string                  // Link type
    Label           string                  // Link label
    Source          string                  // Source node ID
    Target          string                  // Target node ID
}

type RoutingTable struct {
    Default         bool                    // Is default route
    Sequence        int                     // Routing sequence
    Data            string                  // Condition data field
    Value           string                  // Condition value
    Target          string                  // Target node
}
```

### 3.4 Page and View Models (Codegen Support)

**Location:** `/home/user/iac/codegen/pagegen.go`

Pages and views are generated with the following structure:

```go
type PageDefinition struct {
    Name            string                  // Page identifier
    Title           string                  // Display title
    Description     string                  // Page purpose
    Orientation     int                     // Layout orientation
    Version         string                  // Page version
    IsDefault       bool                    // Default version flag
    Status          int                     // Status code
    
    Panels          []PanelDefinition       // UI panels
    Actions         []ActionDefinition      // Page actions
}

type PanelDefinition struct {
    ID              string                  // Panel identifier
    Type            string                  // grid|form|detail|custom
    Title           string                  // Panel title
    View            string                  // Associated view name
    Position        PositionInfo            // Layout position
    Properties      map[string]interface{}  // Panel-specific properties
}

type ViewDefinition struct {
    Name            string                  // View identifier
    Type            string                  // list|form|detail|custom
    Title           string                  // View title
    Datasource      string                  // Table or TranCode name
    Fields          []FieldDefinition       // View fields
    Actions         []ActionDefinition      // View actions
}

type ActionDefinition struct {
    Name            string                  // Action identifier
    Type            string                  // button|link|menu
    Target          string                  // Target name
    TargetType      string                  // view|trancode|url
    Icon            string                  // Font Awesome icon
    Label           string                  // Button label
    Position        string                  // toolbar|context|inline
}
```

### 3.5 Whiteboard Model (Excalidraw Integration)

**Location:** `/home/user/iac/codegen/whiteboardgen.go`

Whiteboard elements are Excalidraw-compatible:

```go
type WhiteboardElement struct {
    ID              string                  // Unique element ID
    Type            string                  // rectangle|ellipse|diamond|arrow|line|text|freedraw
    
    // Positioning
    X, Y            float64                 // Coordinates
    Width, Height   float64                 // Dimensions
    Angle           float64                 // Rotation angle
    
    // Styling
    StrokeColor     string                  // Hex color
    BackgroundColor string                  // Hex color
    FillStyle       string                  // hachure|cross-hatch|solid
    StrokeWidth     int                     // Border thickness
    StrokeStyle     string                  // solid|dashed|dotted
    
    // Effects
    Roughness       int                     // Hand-drawn effect (0-2)
    Opacity         int                     // Transparency (0-100)
    
    // Text-specific
    Text            string                  // Text content
    FontSize        int                     // Font size
    FontFamily      int                     // 1=Virgil, 2=Helvetica, 3=Cascadia
    TextAlign       string                  // left|center|right
    
    // Relationships
    GroupIds        []string                // Group membership
    BoundElements   interface{}             // Arrow bindings
    ContainerId     *string                 // Parent container
    
    // Metadata
    IsDeleted       bool                    // Deletion flag
    Locked          bool                    // Lock status
    Updated         int64                   // Last update timestamp
    Link            *string                 // URL link
}
```

---

## 4. DATA RELATIONSHIPS AND FLOW

### 4.1 Execution Flow Architecture

```
API Request
    ↓
Controller (e.g., TransController)
    ↓
Service Layer (e.g., BusinessEntityService)
    ↓
Database Selection (DatabaseSelector)
    ├─ Primary DB (MySQL/PostgreSQL/MSSQL/Oracle)
    └─ Document DB (MongoDB for configs)
    ↓
TranCode Execution Engine
    ├─ Load TranCode from MongoDB (Transaction_Code collection)
    ├─ Execute Function Groups sequentially
    │   ├─ Functions reference data from relational DB
    │   ├─ Functions update MongoDB document collections
    │   └─ Functions interact with Workflows
    └─ Return results to API response
    ↓
Response to Client
```

### 4.2 Data Integration Points

**TranCode → Relational Database:**
- Query functions execute SQL against configured database
- Insert/Update/Delete functions perform table operations
- Supports multi-database queries with database selector

**TranCode → Document Database (MongoDB):**
- Configuration data stored in collections
- Workflow definitions in WorkFlow collection
- Job history in Job_History collection
- Process plans in Process_Plan collection

**Workflow → TranCode Execution:**
- Workflow nodes reference TranCode names
- Conditional routing based on TranCode output
- Node conditions and postconditions evaluated

**Page/View → TranCode:**
- Pages reference TranCode for data loading
- Views display results from TranCode execution
- Actions trigger TranCode execution

---

## 5. SCHEMA VERSIONING AND DEPLOYMENT MECHANISMS

### 5.1 Version Management System

**Location:** `/home/user/iac/databases/version_manager.go`

The version manager tracks and applies schema migrations:

```go
type VersionManager struct {
    config      *VersionManagerConfig   // Configuration
    db          *sql.DB                 // Database connection
    dbType      string                  // Database type identifier
    migrations  []*Migration             // Registered migrations
    mu          sync.RWMutex            // Thread safety
}

type Migration struct {
    Version     int                     // Migration version number
    Description string                  // Migration description
    UpSQL       string                  // SQL to apply migration
    DownSQL     string                  // SQL to rollback migration
    Checksum    string                  // SHA256 checksum for validation
}

type Version struct {
    Number      int                     // Version number
    Description string                  // Migration description
    AppliedAt   time.Time               // When applied
    Applied     bool                    // Applied flag
    Checksum    string                  // Migration checksum
}
```

**Version Tracking Table:**

The system creates a `schema_versions` table (configurable) to track applied migrations:

- **MySQL:** InnoDB with indexes on version and applied_at
- **PostgreSQL:** Standard table with indexes
- **MSSQL:** MSSQL-specific date handling with GETDATE()
- **Oracle:** Oracle-specific date handling with SYSTIMESTAMP

### 5.2 Migration Management Operations

**Core Operations:**

```go
// Register a new migration
func (vm *VersionManager) RegisterMigration(migration *Migration) error

// Get current database version
func (vm *VersionManager) GetCurrentVersion(ctx context.Context) (int, error)

// Get applied migrations
func (vm *VersionManager) GetAppliedVersions(ctx context.Context) ([]*Version, error)

// Get pending migrations
func (vm *VersionManager) GetPendingMigrations(ctx context.Context) ([]*Migration, error)

// Apply all pending migrations
func (vm *VersionManager) Migrate(ctx context.Context) error

// Migrate to specific version
func (vm *VersionManager) MigrateTo(ctx context.Context, targetVersion int) error

// Migrate up (upgrade)
func (vm *VersionManager) migrateUp(ctx context.Context, current, target int) error

// Migrate down (downgrade/rollback)
func (vm *VersionManager) migrateDown(ctx context.Context, current, target int) error

// Check compatibility
func (vm *VersionManager) CheckCompatibility(ctx context.Context, minVersion, maxVersion int) error

// Get migration history
func (vm *VersionManager) GetMigrationHistory(ctx context.Context) ([]MigrationRecord, error)

// Validate checksum
func (vm *VersionManager) validateChecksum(ctx context.Context, migration *Migration) error
```

### 5.3 Configuration for Version Management

```go
type VersionManagerConfig struct {
    VersionTable        string  // Table name for version tracking (default: schema_versions)
    AutoMigrate         bool    // Auto-migrate on startup (default: false)
    ValidateChecksums   bool    // Validate migration checksums (default: true)
    AllowOutOfOrder     bool    // Allow out-of-order migrations (default: false)
}

// Default configuration
func DefaultVersionManagerConfig() *VersionManagerConfig {
    return &VersionManagerConfig{
        VersionTable:      "schema_versions",
        AutoMigrate:       false,
        ValidateChecksums: true,
        AllowOutOfOrder:   false,
    }
}
```

### 5.4 Docker-Based Deployment

**Location:** `/home/user/iac/docker-compose.databases.yml`

Comprehensive Docker Compose setup for all database types:

**Services:**

1. **MySQL Primary** (port 3306)
   - Image: mysql:8.0
   - Database: iac
   - Auto-initialization with `/docker-entrypoint-initdb.d/init.sql`

2. **MySQL Replica** (port 3307)
   - Replication setup with primary

3. **PostgreSQL** (port 5432)
   - Image: postgres:15
   - JSON/JSONB support
   - Auto-initialization scripts

4. **MSSQL** (port 1433)
   - Image: mssql/server:2022
   - Developer edition

5. **Oracle** (port 1521)
   - Image: gvenzl/oracle-xe:21-slim
   - Express edition

6. **MongoDB** (port 27017)
   - Image: mongo:7.0
   - Authentication enabled
   - Auto-initialization with JavaScript

7. **Redis** (port 6379)
   - Caching support
   - Alpine-based image

8. **Admin Tools:**
   - phpMyAdmin (port 8080) - MySQL administration
   - pgAdmin (port 8081) - PostgreSQL administration
   - Mongo Express (port 8082) - MongoDB administration

### 5.5 Initialization System

**Location:** `/home/user/iac/dbinitializer/initializer.go`

```go
type DatabaseInitializer struct {
    RelationalDBs       map[string]dbconn.RelationalDB
    DocumentDBs         map[string]documents.DocumentDB
    PoolManager         *dbconn.PoolManager
    DocManager          *documents.DocDBManager
    iLog                logger.Log
}

Initialization Process:
1. Load configuration from environment/config files
2. Create primary relational database connection
3. Set up replica databases (optional)
4. Create document database connection
5. Start health monitoring goroutines
6. Apply pending migrations
7. Ready for operation
```

### 5.6 Deployment Scripts

**Location:** `/home/user/iac/scripts/`

Scripts for each database type:

- `mysql/init.sql` - MySQL initialization
- `postgres/init.sql` - PostgreSQL initialization
- `mssql/init.sql` - MSSQL initialization
- `oracle/init.sql` - Oracle initialization
- `mongodb/init.js` - MongoDB initialization

**Script Functions:**
- Create application database and users
- Set up replication (if applicable)
- Create base tables/collections
- Set up indexes
- Configure permissions

---

## 6. CONFIGURATION FILES AND STRUCTURE

### 6.1 Main Configuration File

**Location:** `/home/user/iac/configuration.json`

```json
{
    "instance": "IAC_Instance",
    "name": "IAC",
    "version": "1.0.0",
    
    "database": {
        "type": "mysql",
        "connection": "user:password@tcp(localhost:3306)/iac25?timeout=5s",
        "maxidleconns": 5,
        "maxopenconns": 10,
        "timeout": 5
    },
    
    "altdatabases": [
        {
            "name": "conn1",
            "type": "mysql",
            "connection": "user:password@tcp(localhost:3306)/iac"
        }
    ],
    
    "documentdb": {
        "type": "mongodb",
        "connection": "mongodb://localhost:27017",
        "database": "IAC_CFG"
    },
    
    "webserver": {
        "port": 8000,
        "paths": {
            "portal": {
                "path": "../iac-ui/portal",
                "home": "../iac-ui/portal/uipage.html"
            }
        },
        "proxy": {
            "draw": "http://localhost:3000"
        }
    },
    
    "cache": {
        "adapter": "memory",
        "interval": 7200,
        "objectinterval": 1800,
        "documentdb": {
            "conn": "mongodb://localhost:27017",
            "db": "IAC_CACHE",
            "collection": "cache"
        }
    },
    
    "log": {
        "adapter": "console",
        "level": "debug",
        "performance": true,
        "documentdb": {
            "conn": "mongodb://localhost:27017",
            "db": "IAC_CACHE",
            "collection": "logger"
        }
    }
}
```

### 6.2 Database Configuration Schema

**Relational Database Configuration:**

```go
type DatabaseConfig struct {
    Type                string              // mysql|postgres|mssql|oracle
    Host                string              // Server hostname
    Port                int                 // Connection port
    Database            string              // Database name
    Username            string              // Database user
    Password            string              // Database password
    Connection          string              // Full connection string (optional)
    MaxIdleConns        int                 // Connection pool idle count
    MaxOpenConns        int                 // Connection pool max count
    ConnMaxLifetime     int                 // Connection lifetime in seconds
    ConnMaxIdleTime     int                 // Idle timeout in seconds
    Timeout             int                 // Connection timeout
    SSLMode             string              // SSL configuration
    Options             map[string]string   // Driver-specific options
}
```

**Document Database Configuration:**

```go
type DocumentConfig struct {
    Type                string              // mongodb|postgres
    Host                string              // Server hostname
    Port                int                 // Connection port
    Database            string              // Database name
    Username            string              // Authentication user
    Password            string              // Authentication password
    Connection          string              // Full connection string (optional)
    MaxPoolSize         int                 // Max connection pool size
    MinPoolSize         int                 // Min connection pool size
    Timeout             int                 // Connection timeout
    SSLMode             string              // SSL mode
    AuthSource          string              // MongoDB auth source (admin)
    ReplicaSet          string              // MongoDB replica set name
    Options             map[string]string   // Driver-specific options
}
```

### 6.3 Environment Variable Configuration

The system supports environment variable overrides:

**Relational Database:**
```bash
DB_TYPE=postgres                    # Database type
DB_PRIMARY_HOST=db.example.com      # Primary host
DB_PRIMARY_USER=iac_user            # Username
DB_PRIMARY_PASSWORD=secret          # Password
DB_PRIMARY_NAME=iac                 # Database name
DB_REPLICA_COUNT=2                  # Number of replicas
DB_REPLICA_1_HOST=replica1.example.com
DB_REPLICA_1_USER=iac_user
```

**Document Database:**
```bash
DOCDB_TYPE=mongodb                  # Document DB type
DOCDB_HOST=mongo.example.com        # Host
DOCDB_USER=admin                    # Username
DOCDB_PASSWORD=mongo_pass           # Password
DOCDB_NAME=IAC_CFG                  # Database name
```

---

## 7. ADVANCED FEATURES

### 7.1 Query Caching

**Location:** `/home/user/iac/databases/query_cache.go`

Intelligent query result caching with TTL and invalidation:

- Configurable TTL per query type
- Automatic invalidation on data changes
- LRU eviction policy
- Thread-safe cache operations
- Metrics collection

### 7.2 Replica Manager

**Location:** `/home/user/iac/databases/replica_manager.go`

Read/write splitting and replica load balancing:

- Automatic health checks
- Weight-based load balancing
- Failover on unhealthy replica
- Connection pooling per replica
- Replica selection strategies

### 7.3 Backup & Restore

**Location:** `/home/user/iac/databases/backup_manager.go`

Database backup and restoration capabilities:

```go
type BackupManager struct {
    db              RelationalDB
    backupDir       string
    scheduler       *BackupScheduler
}

Methods:
- CreateBackup()    // Full database backup
- RestoreBackup()   // Restore from backup
- ScheduleBackup()  // Automated backups
- ListBackups()     // Available backups
- VerifyBackup()    // Backup validation
```

### 7.4 Database Security

**Location:** `/home/user/iac/databases/security.go`

Security features including:

- SQL injection prevention (parameterized queries)
- Query logging and audit trails
- Rate limiting
- Connection encryption (SSL/TLS)
- Authentication and authorization

---

## 8. MONITORING AND HEALTH CHECKS

### 8.1 Health Check System

**Location:** `/home/user/iac/health/`

Comprehensive health checks for:

- MySQL database connectivity
- PostgreSQL database connectivity
- MSSQL database connectivity
- MongoDB connectivity
- Connection pool statistics
- Query performance metrics

### 8.2 Metrics Collection

**Location:** `/home/user/iac/metrics/`

Performance metrics:

- Query execution times
- Connection pool usage
- Cache hit rates
- Database availability
- Transaction duration
- Error rates

---

## 9. TESTING INFRASTRUCTURE

### 9.1 Test Setup

**Location:** `/home/user/iac/tests/` and `Makefile.tests`

Comprehensive test suite:

```bash
make test-databases     # Test database layer
make test-integration   # Integration tests
make test-unit          # Unit tests
```

### 9.2 Test Database Setup

Docker-based test environment with all database types:

```bash
scripts/start-databases.sh  # Start all databases
scripts/run-tests.sh        # Run test suite
```

---

## 10. CURRENT IMPLEMENTATION STATUS

### Completed Components:

- Multi-database support architecture
- Configuration management system
- Connection pooling
- Database factory pattern
- Dialect system for SQL differences
- Version management and migrations
- MongoDB document database integration
- Docker deployment setup
- Health monitoring system
- Backup/restore framework

### In Progress:

- Multi-database query builder
- PostgreSQL JSONB document storage
- Advanced caching strategies
- Distributed transaction management
- Sharding support

### Planned Features:

- Connection pooling optimization
- Advanced query optimization
- Real-time replication monitoring
- Multi-region deployment support
- Database clustering

---

## 11. USAGE EXAMPLES

### 11.1 Database Connection

```go
// Get factory instance
factory := dbconn.GetFactory()

// Create relational database
config := &dbconn.DBConfig{
    Type: dbconn.DBTypePostgreSQL,
    Host: "localhost",
    Port: 5432,
    Database: "iac",
    Username: "iac_user",
    Password: "password",
}

db, err := factory.NewRelationalDB(config)
if err != nil {
    log.Fatal(err)
}

// Execute query
rows, err := db.Query(context.Background(), "SELECT * FROM users")
```

### 11.2 Document Database Operations

```go
// Initialize MongoDB
docDB, err := documents.InitMongoDB("mongodb://localhost:27017", "IAC_CFG")
if err != nil {
    log.Fatal(err)
}

// Query collection
filter := bson.M{"trancodename": "SampleTranCode", "isdefault": true}
results, err := docDB.QueryCollection("Transaction_Code", filter, nil)

// Insert document
data := map[string]interface{}{
    "name": "NewTranCode",
    "version": "1.0",
}
result, err := docDB.InsertCollection("Transaction_Code", data)

// Update document
update := bson.M{
    "$set": bson.M{
        "status": "Production",
    },
}
err = docDB.UpdateCollection("Transaction_Code", filter, update, nil)
```

### 11.3 Migration Management

```go
// Create version manager
vm, err := dbconn.NewVersionManager(db, "postgres", nil)
if err != nil {
    log.Fatal(err)
}

// Register migration
vm.RegisterMigration(&dbconn.Migration{
    Version: 1,
    Description: "Create users table",
    UpSQL: `CREATE TABLE users (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        email VARCHAR(255) UNIQUE NOT NULL
    )`,
    DownSQL: "DROP TABLE users",
})

// Apply migrations
err = vm.Migrate(context.Background())
```

---

## 12. ARCHITECTURAL DIAGRAM

```
┌─────────────────────────────────────────────────────────────────┐
│                          API Layer                               │
│                (Controllers & Endpoints)                        │
└──────────────────────────┬──────────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────────┐
│                      Service Layer                              │
│   (BusinessEntityService, DatabaseHelper, Factories)           │
└──────────────────────────┬──────────────────────────────────────┘
                           │
        ┌──────────────────┴──────────────────┐
        │                                      │
┌───────▼─────────────────────────┐  ┌────────▼────────────────────────┐
│   Database Selector/Router       │  │   Cache Manager                │
│   - Primary/Replica selection    │  │   - Memory/Redis/File         │
│   - Load balancing               │  │   - TTL management            │
└───────┬─────────────────────────┘  └────────────────────────────────┘
        │
        ├─────────────────────┬──────────────────┬──────────────────┐
        │                      │                  │                  │
    ┌───▼────────┐  ┌────────▼─────┐  ┌─────────▼────┐  ┌─────────▼────┐
    │   MySQL     │  │ PostgreSQL   │  │    MSSQL     │  │    Oracle    │
    │  Dialect    │  │  Dialect     │  │   Dialect    │  │   Dialect    │
    │   Pool      │  │   Pool       │  │    Pool      │  │    Pool      │
    └─────────────┘  └──────────────┘  └──────────────┘  └──────────────┘
        │                  │                  │               │
        └──────────────────┼──────────────────┼───────────────┘
                           │
                ┌──────────▼──────────┐
                │  Connection Pooling │
                │  - Pool per DB type │
                │  - Health checks    │
                │  - Metrics tracking │
                └─────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    Document Database Layer                       │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │             MongoDB Interface                            │  │
│  │  - QueryCollection, InsertCollection, UpdateCollection │  │
│  │  - DeleteItemFromCollection                             │  │
│  │  - Connection pooling & monitoring                      │  │
│  └──────────────────────────────────────────────────────────┘  │
│                           │                                     │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Collections:                                            │  │
│  │  - Transaction_Code (TranCode definitions)             │  │
│  │  - WorkFlow (Workflow orchestration)                    │  │
│  │  - Process_Plan, Job_History, Notifications            │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    Engine & Execution                            │
│                                                                  │
│  ┌────────────────────────────────────────────────────────┐    │
│  │  TranCode Executor                                     │    │
│  │  - Function Group sequencing                          │    │
│  │  - Input/Output mapping                               │    │
│  │  - Transaction management                             │    │
│  │  - Error handling                                      │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                  │
│  ┌────────────────────────────────────────────────────────┐    │
│  │  Workflow Engine                                       │    │
│  │  - Node execution                                      │    │
│  │  - Conditional routing                                │    │
│  │  - Integration with TranCode                          │    │
│  └────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                Version Management & Deployment                   │
│                                                                  │
│  ┌────────────────────────────────────────────────────────┐    │
│  │  Migration Manager                                     │    │
│  │  - Schema versioning                                  │    │
│  │  - Migration tracking (schema_versions table)         │    │
│  │  - Rollback support                                    │    │
│  │  - Multi-DB dialect support                           │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                  │
│  ┌────────────────────────────────────────────────────────┐    │
│  │  Docker Deployment                                    │    │
│  │  - Multi-database compose files                       │    │
│  │  - Automated initialization                           │    │
│  │  - Admin tools (phpMyAdmin, pgAdmin, Mongo Express)  │    │
│  └────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

---

## 13. KEY FILES REFERENCE

| Component | Location | Purpose |
|-----------|----------|---------|
| Relational DB Interface | `databases/interface.go` | Abstract DB operations |
| Database Factory | `databases/factory.go` | Driver instantiation |
| DB Operations | `databases/dboperation.go` | Query execution |
| Connection Pool | `databases/poolmanager.go` | Connection management |
| Version Manager | `databases/version_manager.go` | Schema migrations |
| Document DB Interface | `documents/interface.go` | Abstract document ops |
| MongoDB Implementation | `documents/mongodb.go` | MongoDB integration |
| TranCode Types | `engine/types/types.go` | Data models |
| Workflow Types | `workflow/types/type.go` | Workflow structures |
| Codegen | `codegen/` | Code generation for UIs |
| Configuration | `config/` | Config loading & management |
| Database Initializer | `dbinitializer/initializer.go` | Initialization process |
| Docker Setup | `docker-compose.databases.yml` | Deployment containers |
| Scripts | `scripts/` | DB-specific init scripts |

---

## 14. CONCLUSION

The IAC framework implements a comprehensive, production-ready database architecture that:

1. **Supports Multiple Database Types:** MySQL, PostgreSQL, MSSQL, Oracle for relational data, MongoDB for document data
2. **Provides Abstraction:** Interface-based design allows database-agnostic application code
3. **Manages Complexity:** Factory patterns, connection pooling, and dialect system handle vendor differences
4. **Enables Deployment:** Docker-based setup with automated initialization and health checks
5. **Tracks Changes:** Version management system with migration support and rollback capability
6. **Monitors Health:** Built-in health checks, metrics collection, and connection monitoring
7. **Ensures Performance:** Query caching, connection pooling, read/write splitting
8. **Maintains Security:** Parameterized queries, audit logging, encryption support

This architecture enables building enterprise applications with flexibility to choose their database backend while maintaining consistent application interfaces and operations.

