-- IAC MS SQL Server Initialization Script
-- This script creates the basic schema for IAC

-- Create database
IF NOT EXISTS (SELECT * FROM sys.databases WHERE name = 'iac')
BEGIN
    CREATE DATABASE iac;
END
GO

USE iac;
GO

-- Create sample tables for testing
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'users')
BEGIN
    CREATE TABLE users (
        id INT IDENTITY(1,1) PRIMARY KEY,
        uuid UNIQUEIDENTIFIER DEFAULT NEWID() UNIQUE NOT NULL,
        username NVARCHAR(50) UNIQUE NOT NULL,
        email NVARCHAR(100) UNIQUE NOT NULL,
        password_hash NVARCHAR(255) NOT NULL,
        first_name NVARCHAR(50),
        last_name NVARCHAR(50),
        status NVARCHAR(20) DEFAULT 'pending' CHECK (status IN ('active', 'inactive', 'pending')),
        created_at DATETIME2 DEFAULT GETUTCDATE(),
        updated_at DATETIME2 DEFAULT GETUTCDATE()
    );

    CREATE INDEX idx_users_username ON users(username);
    CREATE INDEX idx_users_email ON users(email);
    CREATE INDEX idx_users_status ON users(status);
END
GO

-- Create sessions table
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'sessions')
BEGIN
    CREATE TABLE sessions (
        id INT IDENTITY(1,1) PRIMARY KEY,
        session_id NVARCHAR(64) UNIQUE NOT NULL,
        user_id INT NOT NULL,
        ip_address NVARCHAR(45),
        user_agent NVARCHAR(MAX),
        last_activity DATETIME2 DEFAULT GETUTCDATE(),
        created_at DATETIME2 DEFAULT GETUTCDATE(),
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    );

    CREATE INDEX idx_sessions_session_id ON sessions(session_id);
    CREATE INDEX idx_sessions_user_id ON sessions(user_id);
    CREATE INDEX idx_sessions_last_activity ON sessions(last_activity);
END
GO

-- Create audit log table
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'audit_log')
BEGIN
    CREATE TABLE audit_log (
        id BIGINT IDENTITY(1,1) PRIMARY KEY,
        user_id INT,
        action NVARCHAR(50) NOT NULL,
        entity_type NVARCHAR(50),
        entity_id NVARCHAR(100),
        changes NVARCHAR(MAX), -- JSON data
        ip_address NVARCHAR(45),
        created_at DATETIME2 DEFAULT GETUTCDATE()
    );

    CREATE INDEX idx_audit_log_user_id ON audit_log(user_id);
    CREATE INDEX idx_audit_log_action ON audit_log(action);
    CREATE INDEX idx_audit_log_created_at ON audit_log(created_at);
END
GO

-- Insert sample data
IF NOT EXISTS (SELECT * FROM users WHERE username = 'admin')
BEGIN
    INSERT INTO users (uuid, username, email, password_hash, first_name, last_name, status) VALUES
    ('550E8400-E29B-41D4-A716-446655440000', 'admin', 'admin@iac.local', '$2a$10$rZfE8qvd1xqY.T9hG3V8H.', 'Admin', 'User', 'active'),
    ('660E8400-E29B-41D4-A716-446655440001', 'testuser', 'test@iac.local', '$2a$10$rZfE8qvd1xqY.T9hG3V8H.', 'Test', 'User', 'active');
END
GO

-- Display success message
SELECT 'IAC MS SQL Server database initialized successfully' AS message;
GO
