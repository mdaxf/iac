-- Agent Runtime v4: Add mcp_path to mcp_servers (standard MCP protocol endpoint path)
-- Run after agent_runtime_v3_postgresql.sql

ALTER TABLE mcp_servers
  ADD COLUMN IF NOT EXISTS mcp_path VARCHAR(200) NOT NULL DEFAULT '/mcp';
