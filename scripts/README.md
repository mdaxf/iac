# IAC Database Setup Scripts

This directory contains initialization scripts for all supported databases in the IAC system.

## Quick Start

### Start All Databases

```bash
# Start all databases using Docker Compose
docker-compose -f docker-compose.databases.yml up -d

# Check status
docker-compose -f docker-compose.databases.yml ps

# View logs
docker-compose -f docker-compose.databases.yml logs -f
```

### Start Individual Databases

```bash
# MySQL only
docker-compose -f docker-compose.databases.yml up -d mysql-primary

# PostgreSQL only
docker-compose -f docker-compose.databases.yml up -d postgres

# MongoDB only
docker-compose -f docker-compose.databases.yml up -d mongodb

# MSSQL only
docker-compose -f docker-compose.databases.yml up -d mssql

# Oracle only
docker-compose -f docker-compose.databases.yml up -d oracle
```

### Stop Databases

```bash
# Stop all
docker-compose -f docker-compose.databases.yml down

# Stop and remove volumes (WARNING: This deletes all data)
docker-compose -f docker-compose.databases.yml down -v
```

## Database Administration Tools

Access the web-based admin tools:

- **phpMyAdmin** (MySQL): http://localhost:8080
- **pgAdmin** (PostgreSQL): http://localhost:8081
  - Email: admin@iac.local
  - Password: pgadmin_pass
- **Mongo Express** (MongoDB): http://localhost:8082
  - Username: admin
  - Password: admin

## Connection Information

### MySQL

```
Host: localhost
Port: 3306
Database: iac
Username: iac_user
Password: iac_pass
Root Password: mysql_root_pass
```

Connection String:
```
mysql://iac_user:iac_pass@localhost:3306/iac
```

### PostgreSQL

```
Host: localhost
Port: 5432
Database: iac
Username: iac_user
Password: iac_pass
```

Connection String:
```
postgresql://iac_user:iac_pass@localhost:5432/iac
```

### Microsoft SQL Server

```
Host: localhost
Port: 1433
Database: iac
Username: sa
Password: MsSql_Pass123!
```

Connection String:
```
sqlserver://sa:MsSql_Pass123!@localhost:1433?database=iac
```

### Oracle

```
Host: localhost
Port: 1521
SID: xe
Service Name: iac
Username: iac_user
Password: iac_pass
System Password: Oracle_Pass123
```

Connection String:
```
oracle://iac_user:iac_pass@localhost:1521/iac
```

### MongoDB

```
Host: localhost
Port: 27017
Database: iac
Username: iac_user
Password: iac_pass
Admin Username: admin
Admin Password: mongo_pass
```

Connection String:
```
mongodb://iac_user:iac_pass@localhost:27017/iac?authSource=iac
```

## Initialization Scripts

Each database has its own initialization script that runs automatically when the container starts for the first time:

- **MySQL**: `scripts/mysql/init.sql`
- **PostgreSQL**: `scripts/postgres/init.sql`
- **MongoDB**: `scripts/mongodb/init.js`
- **MSSQL**: `scripts/mssql/init.sql`
- **Oracle**: `scripts/oracle/init.sql`

These scripts create:
- Sample tables (`users`, `sessions`, `audit_log`)
- Indexes for performance
- Sample data for testing
- Required permissions

## Testing Database Connections

### Using Docker

```bash
# MySQL
docker exec -it iac-mysql-primary mysql -u iac_user -piac_pass iac

# PostgreSQL
docker exec -it iac-postgres psql -U iac_user -d iac

# MongoDB
docker exec -it iac-mongodb mongosh iac -u iac_user -p iac_pass

# MSSQL
docker exec -it iac-mssql /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P MsSql_Pass123!

# Oracle
docker exec -it iac-oracle sqlplus iac_user/iac_pass@localhost:1521/iac
```

### Using IAC Application

Set environment variables in `.env`:

```bash
# For MySQL
DB_TYPE=mysql
DB_HOST=localhost
DB_PORT=3306
DB_DATABASE=iac
DB_USERNAME=iac_user
DB_PASSWORD=iac_pass

# For PostgreSQL
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_DATABASE=iac
DB_USERNAME=iac_user
DB_PASSWORD=iac_pass

# For MongoDB (Document DB)
DOCDB_TYPE=mongodb
DOCDB_HOST=localhost
DOCDB_PORT=27017
DOCDB_DATABASE=iac
DOCDB_USERNAME=iac_user
DOCDB_PASSWORD=iac_pass
```

## Troubleshooting

### Database won't start

```bash
# Check logs
docker-compose -f docker-compose.databases.yml logs [service-name]

# Example for MySQL
docker-compose -f docker-compose.databases.yml logs mysql-primary
```

### Reset database

```bash
# Stop and remove containers and volumes
docker-compose -f docker-compose.databases.yml down -v

# Start fresh
docker-compose -f docker-compose.databases.yml up -d
```

### Permission errors

For Oracle:
```bash
# Grant DBA role if needed
docker exec -it iac-oracle sqlplus system/Oracle_Pass123@localhost:1521/xe
GRANT DBA TO iac_user;
```

For MSSQL:
```bash
# Grant additional permissions if needed
docker exec -it iac-mssql /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P MsSql_Pass123!
USE iac;
ALTER ROLE db_owner ADD MEMBER iac_user;
GO
```

## Production Considerations

**Important**: These scripts and configurations are for development and testing only!

For production deployments:

1. Use strong, unique passwords
2. Enable SSL/TLS connections
3. Configure proper firewall rules
4. Use separate hosts for different databases
5. Implement backup strategies
6. Configure replication
7. Monitor database performance
8. Implement proper access controls
9. Regular security audits
10. Keep databases updated

## Data Persistence

All data is stored in Docker volumes:

- `mysql-primary-data`
- `postgres-data`
- `mongodb-data`
- `mssql-data`
- `oracle-data`
- `redis-data`

To backup volumes:
```bash
docker volume ls
docker run --rm -v mysql-primary-data:/data -v $(pwd):/backup alpine tar czf /backup/mysql-backup.tar.gz /data
```

To restore volumes:
```bash
docker run --rm -v mysql-primary-data:/data -v $(pwd):/backup alpine tar xzf /backup/mysql-backup.tar.gz -C /
```

## Additional Resources

- [MySQL Documentation](https://dev.mysql.com/doc/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [MongoDB Documentation](https://docs.mongodb.com/)
- [MSSQL Documentation](https://docs.microsoft.com/en-us/sql/)
- [Oracle Documentation](https://docs.oracle.com/en/database/)
