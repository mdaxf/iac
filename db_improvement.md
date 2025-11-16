# Database Multi-Support Improvement Plan

**Version:** 1.0
**Date:** 2025-11-15
**Objective:** Extend IAC system to support multiple database types for both relational and document databases

---

## Executive Summary

This document outlines the implementation plan to enhance IAC's database layer to support multiple database types:
- **Relational Databases**: MySQL, PostgreSQL, MSSQL, Oracle
- **Document Databases**: MongoDB, PostgreSQL (JSONB)

**Current State:**
- Relational DB: MySQL and MSSQL only (hardcoded)
- Document DB: MongoDB only
- Hardcoded connection strings in `databases/dbconn.go`
- MySQL-specific monitoring function
- No database abstraction layer

**Target State:**
- Configuration-driven database selection
- Support for 4 relational DB types + 2 document DB types
- Database abstraction layer for vendor-specific operations
- Unified monitoring and health check system
- Connection pooling per database type

---

## Phase 1: Architecture Design & Configuration (8 tasks)

### Task 1.1: Design Database Abstraction Layer
**Priority:** High
**Effort:** 3 days
**Description:** Create interface-based abstraction layer for database operations

**Deliverables:**
- `databases/interface.go` - Define `RelationalDB` interface
- `documents/interface.go` - Define `DocumentDB` interface
- Interface methods:
  ```go
  type RelationalDB interface {
      Connect(config DBConfig) error
      Ping() error
      Query(sql string, args ...interface{}) (*sql.Rows, error)
      Exec(sql string, args ...interface{}) (sql.Result, error)
      BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
      Close() error
      GetDialect() string
      SupportsFeature(feature string) bool
  }
  ```

### Task 1.2: Create Database Configuration Structure
**Priority:** High
**Effort:** 2 days
**Description:** Define configuration schema for all supported databases

**Deliverables:**
- Update `config` package with `DBConfig` struct:
  ```go
  type DBConfig struct {
      Type         string // mysql, postgres, mssql, oracle, mongodb
      Host         string
      Port         int
      Database     string
      Username     string
      Password     string
      SSLMode      string
      MaxIdleConns int
      MaxOpenConns int
      ConnTimeout  int
      Options      map[string]string // DB-specific options
  }
  ```
- Environment variable mapping
- Configuration file (YAML/JSON) schema

### Task 1.3: Implement Configuration Loader
**Priority:** High
**Effort:** 2 days
**Description:** Load database configuration from environment variables and config files

**Deliverables:**
- `config/dbloader.go` - Load from environment variables
- Support for multiple database configs (primary, replica, document)
- Validation and defaults
- Example configuration files

### Task 1.4: Create Database Factory Pattern
**Priority:** High
**Effort:** 2 days
**Description:** Implement factory to instantiate correct database driver

**Deliverables:**
- `databases/factory.go`:
  ```go
  func NewRelationalDB(config DBConfig) (RelationalDB, error)
  func NewDocumentDB(config DBConfig) (DocumentDB, error)
  ```
- Driver registration mechanism
- Connection string builders per database type

### Task 1.5: Design Connection Pool Manager
**Priority:** Medium
**Effort:** 3 days
**Description:** Manage multiple database connection pools

**Deliverables:**
- `databases/poolmanager.go`
- Support for multiple named connections
- Connection lifecycle management
- Pool statistics and monitoring

### Task 1.6: Create Database Dialect System
**Priority:** Medium
**Effort:** 2 days
**Description:** Handle SQL dialect differences across databases

**Deliverables:**
- `databases/dialect/dialect.go` - Base dialect interface
- Dialect-specific query builders
- Data type mapping tables
- Pagination syntax differences

### Task 1.7: Design Migration System for Multi-DB
**Priority:** Medium
**Effort:** 3 days
**Description:** Support schema migrations across different database types

**Deliverables:**
- Database-specific migration files
- Migration version tracking per DB type
- Rollback support
- Migration testing framework

### Task 1.8: Update GORM Bridge for Multi-DB
**Priority:** Medium
**Effort:** 2 days
**Description:** Extend `gormdb` package to support multiple databases

**Deliverables:**
- Update `gormdb/gormdb.go` to accept database type
- GORM driver selection based on DB type
- Dialect configuration for GORM

---

## Phase 2: Relational Database Drivers (12 tasks)

### Task 2.1: Implement MySQL Driver Adapter
**Priority:** High
**Effort:** 2 days
**Description:** Refactor existing MySQL code into adapter pattern

**Files to Create/Modify:**
- `databases/mysql/mysql.go` - MySQL adapter implementation
- `databases/mysql/dialect.go` - MySQL dialect
- `databases/mysql/monitor.go` - Refactor existing monitor

**Dependencies:**
- `github.com/go-sql-driver/mysql`

### Task 2.2: Implement PostgreSQL Driver Adapter
**Priority:** High
**Effort:** 3 days
**Description:** Add PostgreSQL support for relational database

**Files to Create:**
- `databases/postgres/postgres.go` - PostgreSQL adapter
- `databases/postgres/dialect.go` - PostgreSQL dialect
- `databases/postgres/monitor.go` - Connection monitoring

**Dependencies:**
- `github.com/lib/pq` or `github.com/jackc/pgx/v5`

**PostgreSQL-Specific Features:**
- LISTEN/NOTIFY support
- JSONB column support
- Array type handling
- Full-text search

### Task 2.3: Implement MSSQL Driver Adapter
**Priority:** High
**Effort:** 3 days
**Description:** Refactor existing MSSQL code into adapter pattern

**Files to Create/Modify:**
- `databases/mssql/mssql.go` - MSSQL adapter
- `databases/mssql/dialect.go` - T-SQL dialect
- `databases/mssql/monitor.go` - Connection monitoring

**Dependencies:**
- `github.com/denisenkom/go-mssqldb` (already imported)

**MSSQL-Specific Features:**
- Windows authentication support
- Stored procedure handling
- Table-valued parameters

### Task 2.4: Implement Oracle Driver Adapter
**Priority:** Medium
**Effort:** 5 days
**Description:** Add Oracle database support

**Files to Create:**
- `databases/oracle/oracle.go` - Oracle adapter
- `databases/oracle/dialect.go` - Oracle PL/SQL dialect
- `databases/oracle/monitor.go` - Connection monitoring

**Dependencies:**
- `github.com/sijms/go-ora/v2` or `github.com/godror/godror`

**Oracle-Specific Features:**
- Sequence handling
- ROWNUM vs ROWID
- PL/SQL block execution
- LOB handling

### Task 2.5: Create SQL Translator Utility
**Priority:** Medium
**Effort:** 4 days
**Description:** Translate SQL queries between database dialects

**Deliverables:**
- `databases/translator/translator.go`
- Common SQL → DB-specific SQL
- Pagination translation (LIMIT vs TOP vs ROWNUM)
- Date function translation

### Task 2.6: Implement Transaction Manager
**Priority:** High
**Effort:** 3 days
**Description:** Unified transaction handling across databases

**Deliverables:**
- `databases/transaction/manager.go`
- Nested transaction support
- Savepoint handling
- Two-phase commit preparation

### Task 2.7: Add Connection Health Checks
**Priority:** High
**Effort:** 2 days
**Description:** Health check implementation for each database type

**Files to Create/Modify:**
- Update `health/checks/` directory
- `health/checks/mysql.go` (already exists - enhance)
- `health/checks/postgres.go`
- `health/checks/mssql.go`
- `health/checks/oracle.go`

### Task 2.8: Implement Auto-Reconnection Logic
**Priority:** High
**Effort:** 3 days
**Description:** Generic reconnection for all database types

**Deliverables:**
- Refactor `monitorAndReconnectMySQL` to be generic
- `databases/reconnect/reconnect.go`
- Exponential backoff strategy
- Circuit breaker pattern

### Task 2.9: Add Query Logging and Metrics
**Priority:** Medium
**Effort:** 2 days
**Description:** Database query logging across all database types

**Deliverables:**
- Query execution time tracking
- Slow query logging
- Database metrics collection
- Integration with existing logger

### Task 2.10: Create Connection String Builders
**Priority:** Medium
**Effort:** 2 days
**Description:** Build connection strings for each database type

**Deliverables:**
- `databases/connstring/builder.go`
- Secure credential handling
- SSL/TLS configuration
- Connection option builders

### Task 2.11: Implement Database-Specific Error Handling
**Priority:** Medium
**Effort:** 3 days
**Description:** Translate database-specific errors to common errors

**Deliverables:**
- `databases/errors/errors.go`
- Error code mapping
- Common error types (Duplicate, NotFound, ConnectionLost, etc.)
- Retry logic for transient errors

### Task 2.12: Add Database Feature Detection
**Priority:** Low
**Effort:** 2 days
**Description:** Detect supported features per database

**Deliverables:**
- `databases/features/detector.go`
- Feature flags (CTEs, Window Functions, JSON, Full-text, etc.)
- Version-based capability detection

---

## Phase 3: Document Database Support (8 tasks)

### Task 3.1: Refactor MongoDB Implementation
**Priority:** High
**Effort:** 2 days
**Description:** Refactor existing MongoDB code to use interface

**Files to Modify:**
- `documents/mongodb.go` - Implement DocumentDB interface
- `framework/documentdb/mongodb/mongodb.go` - Consolidate code

### Task 3.2: Implement PostgreSQL JSONB Document Store
**Priority:** Medium
**Effort:** 4 days
**Description:** Use PostgreSQL as document database with JSONB

**Files to Create:**
- `documents/postgres/postgres.go` - PostgreSQL document adapter
- `documents/postgres/jsonb.go` - JSONB operations
- `documents/postgres/indexing.go` - GIN/GiST indexes

**Features:**
- JSONB query operations
- Document indexing
- Full-text search on documents
- Aggregation using SQL

### Task 3.3: Create Document Database Factory
**Priority:** High
**Effort:** 2 days
**Description:** Factory pattern for document databases

**Deliverables:**
- Update `documents/documentdb.go`
- Support MongoDB and PostgreSQL selection
- Connection management

### Task 3.4: Implement Document Query Builder
**Priority:** Medium
**Effort:** 3 days
**Description:** Unified query interface for document operations

**Deliverables:**
- `documents/query/builder.go`
- MongoDB aggregation pipeline builder
- PostgreSQL JSONB query builder
- Filter, projection, sorting, pagination

### Task 3.5: Add Document Indexing Support
**Priority:** Medium
**Effort:** 2 days
**Description:** Index management for document databases

**Deliverables:**
- MongoDB index creation
- PostgreSQL GIN/GiST indexes
- Index recommendation engine

### Task 3.6: Implement Document Validation
**Priority:** Low
**Effort:** 2 days
**Description:** Schema validation for document databases

**Deliverables:**
- JSON Schema validation
- MongoDB validator integration
- PostgreSQL constraint-based validation

### Task 3.7: Add Document Health Checks
**Priority:** High
**Effort:** 1 day
**Description:** Health checks for document databases

**Files to Create/Modify:**
- Update `health/checks/mongodb.go` (already exists)
- `health/checks/postgres_docdb.go`

### Task 3.8: Create Document Migration Tool
**Priority:** Medium
**Effort:** 3 days
**Description:** Migrate documents between database types

**Deliverables:**
- `documents/migration/migrator.go`
- MongoDB → PostgreSQL migration
- Data transformation pipeline
- Progress tracking

---

## Phase 4: Integration & Testing (10 tasks)

### Task 4.1: Update Main Application Initialization
**Priority:** High
**Effort:** 2 days
**Description:** Modify `main.go` to use new database layer

**Changes:**
- Load database configuration
- Initialize database factory
- Setup connection pools
- Initialize GORM with selected database

### Task 4.2: Update Service Layer for Multi-DB
**Priority:** High
**Effort:** 3 days
**Description:** Update services to work with multiple databases

**Files to Modify:**
- All service files using database
- Use dialect-aware queries
- Handle database-specific features

### Task 4.3: Create Database Selection API
**Priority:** Medium
**Effort:** 2 days
**Description:** Runtime database selection for different operations

**Deliverables:**
- Context-based database selection
- Per-request database routing
- Read replica support

### Task 4.4: Implement Database Switching Tests
**Priority:** High
**Effort:** 3 days
**Description:** Integration tests for each database type

**Deliverables:**
- Test suite for MySQL
- Test suite for PostgreSQL
- Test suite for MSSQL
- Test suite for Oracle
- Test suite for MongoDB
- Test suite for PostgreSQL JSONB

### Task 4.5: Add Database Performance Benchmarks
**Priority:** Medium
**Effort:** 2 days
**Description:** Performance testing across database types

**Deliverables:**
- Benchmark suite
- Performance comparison report
- Optimization recommendations

### Task 4.6: Create Database Setup Scripts
**Priority:** Medium
**Effort:** 2 days
**Description:** Docker Compose and setup scripts for all databases

**Deliverables:**
- `docker-compose.yml` with all databases
- Initialization scripts per database
- Sample data loading scripts

### Task 4.7: Update Documentation
**Priority:** High
**Effort:** 3 days
**Description:** Documentation for multi-database support

**Deliverables:**
- Configuration guide
- Database selection guide
- Migration guide
- Troubleshooting guide
- API documentation updates

### Task 4.8: Create Database Admin CLI Tool
**Priority:** Low
**Effort:** 3 days
**Description:** Command-line tool for database management

**Deliverables:**
- Database connection testing
- Migration execution
- Health check runner
- Backup/restore utilities

### Task 4.9: Implement Database Metrics Dashboard
**Priority:** Low
**Effort:** 3 days
**Description:** Monitoring dashboard for database connections

**Deliverables:**
- Connection pool metrics
- Query performance metrics
- Error rate tracking
- Database health status

### Task 4.10: Security Audit and Hardening
**Priority:** High
**Effort:** 2 days
**Description:** Security review of database layer

**Deliverables:**
- SQL injection prevention review
- Credential management review
- TLS/SSL enforcement
- Access control validation

---

## Phase 5: Advanced Features (6 tasks)

### Task 5.1: Implement Read Replica Support
**Priority:** Medium
**Effort:** 3 days
**Description:** Read/write splitting for scalability

**Deliverables:**
- Primary/replica connection management
- Automatic read routing
- Replica lag monitoring
- Failover handling

### Task 5.2: Add Database Sharding Support
**Priority:** Low
**Effort:** 5 days
**Description:** Horizontal sharding for large datasets

**Deliverables:**
- Shard key selection
- Query routing to shards
- Shard rebalancing
- Cross-shard queries

### Task 5.3: Implement Query Caching Layer
**Priority:** Medium
**Effort:** 3 days
**Description:** Cache frequently accessed queries

**Deliverables:**
- Redis integration for query cache
- Cache invalidation strategy
- Cache hit/miss metrics

### Task 5.4: Add Database Backup/Restore
**Priority:** Medium
**Effort:** 4 days
**Description:** Automated backup and restore

**Deliverables:**
- Scheduled backup jobs
- Point-in-time recovery
- Cross-database backup format
- Restore verification

### Task 5.5: Implement Database Versioning
**Priority:** Low
**Effort:** 2 days
**Description:** Track database schema versions

**Deliverables:**
- Version tracking table
- Schema diff detection
- Compatibility checking

### Task 5.6: Create Database Proxy Layer
**Priority:** Low
**Effort:** 4 days
**Description:** Transparent database proxy for advanced features

**Deliverables:**
- Connection pooling proxy
- Query rewriting proxy
- Monitoring and logging proxy
- Load balancing proxy

---

## Implementation Timeline

### Sprint 1 (Weeks 1-2): Foundation
- Phase 1: Tasks 1.1 - 1.4 (Database abstraction and configuration)

### Sprint 2 (Weeks 3-4): Core Drivers Part 1
- Phase 1: Tasks 1.5 - 1.8 (Pool management, dialects, migration)
- Phase 2: Tasks 2.1 - 2.2 (MySQL and PostgreSQL adapters)

### Sprint 3 (Weeks 5-6): Core Drivers Part 2
- Phase 2: Tasks 2.3 - 2.6 (MSSQL, Oracle, SQL translator, transactions)

### Sprint 4 (Weeks 7-8): Reliability Features
- Phase 2: Tasks 2.7 - 2.12 (Health checks, reconnection, logging, errors)

### Sprint 5 (Weeks 9-10): Document Databases
- Phase 3: Tasks 3.1 - 3.5 (MongoDB refactor, PostgreSQL JSONB, querying)

### Sprint 6 (Weeks 11-12): Document DB Features
- Phase 3: Tasks 3.6 - 3.8 (Validation, health checks, migration)

### Sprint 7 (Weeks 13-14): Integration Part 1
- Phase 4: Tasks 4.1 - 4.5 (App integration, testing, benchmarking)

### Sprint 8 (Weeks 15-16): Integration Part 2
- Phase 4: Tasks 4.6 - 4.10 (Setup scripts, documentation, security)

### Sprint 9 (Weeks 17-18): Advanced Features
- Phase 5: Tasks 5.1 - 5.3 (Replicas, sharding, caching)

### Sprint 10 (Weeks 19-20): Polish & Release
- Phase 5: Tasks 5.4 - 5.6 (Backup, versioning, proxy)
- Final testing and documentation
- Production readiness review

**Total Duration:** 20 weeks (5 months)

---

## Task Summary

| Phase | Tasks | Estimated Effort |
|-------|-------|------------------|
| Phase 1: Architecture | 8 tasks | 19 days |
| Phase 2: Relational DB | 12 tasks | 34 days |
| Phase 3: Document DB | 8 tasks | 19 days |
| Phase 4: Integration | 10 tasks | 24 days |
| Phase 5: Advanced | 6 tasks | 21 days |
| **TOTAL** | **54 tasks** | **117 days (23.4 weeks)** |

---

## Dependencies and Prerequisites

### External Libraries Required:
```go
// Relational Databases
"github.com/go-sql-driver/mysql"           // MySQL
"github.com/lib/pq"                        // PostgreSQL
"github.com/jackc/pgx/v5"                  // PostgreSQL (alternative)
"github.com/denisenkom/go-mssqldb"        // MSSQL (already included)
"github.com/sijms/go-ora/v2"              // Oracle

// Document Databases
"go.mongodb.org/mongo-driver"              // MongoDB (already included)

// ORM Support
"gorm.io/driver/mysql"                     // GORM MySQL
"gorm.io/driver/postgres"                  // GORM PostgreSQL
"gorm.io/driver/sqlserver"                 // GORM MSSQL
"gorm.io/driver/oracle"                    // GORM Oracle (if available)
```

### Infrastructure Requirements:
- Docker containers for all database types (development/testing)
- CI/CD pipeline updates for multi-database testing
- Database credentials management (secrets manager)

---

## Risk Assessment

### High Risk:
1. **Oracle Driver Compatibility** - Oracle drivers may have licensing or compatibility issues
   - Mitigation: Evaluate open-source drivers early, consider commercial options

2. **Breaking Changes** - Refactoring may break existing functionality
   - Mitigation: Comprehensive test coverage before refactoring

3. **Performance Impact** - Abstraction layer may impact performance
   - Mitigation: Benchmark early, optimize hot paths

### Medium Risk:
1. **SQL Dialect Differences** - Complex queries may not translate well
   - Mitigation: Maintain database-specific query overrides

2. **Migration Complexity** - Existing data migration is complex
   - Mitigation: Gradual migration strategy, rollback plans

### Low Risk:
1. **Documentation Lag** - Documentation may fall behind implementation
   - Mitigation: Documentation as part of definition of done

---

## Success Criteria

1. ✅ All 4 relational databases (MySQL, PostgreSQL, MSSQL, Oracle) fully supported
2. ✅ Both document databases (MongoDB, PostgreSQL JSONB) fully supported
3. ✅ Zero breaking changes to existing functionality
4. ✅ 95%+ test coverage for database layer
5. ✅ Less than 5% performance degradation compared to current MySQL-only implementation
6. ✅ Comprehensive documentation for all database types
7. ✅ Successful production deployment with at least 2 database types
8. ✅ Automated health checks and monitoring for all databases
9. ✅ Migration path from current architecture documented and tested
10. ✅ Security audit passed with zero critical issues

---

## Open Questions

1. Should we support database-agnostic migrations or maintain separate migration files per database?
2. What is the default database type for new installations?
3. Do we need to support switching databases after initial deployment?
4. Should document databases support relational operations (joins, etc.)?
5. What is the backward compatibility requirement for existing installations?
6. Should we support multi-tenancy with different database types per tenant?

---

## Next Steps

1. Review and approve this plan with stakeholders
2. Prioritize phases based on business requirements
3. Setup development environment with all databases
4. Begin Phase 1 Task 1.1: Design Database Abstraction Layer
5. Schedule weekly progress reviews
6. Update plan based on learnings and feedback

---

**Document Owner:** Development Team
**Last Updated:** 2025-11-15
**Status:** Draft - Pending Approval
