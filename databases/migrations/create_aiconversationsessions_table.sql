-- Migration: Create AI Conversation Sessions table
-- Description: Stores conversation session state for AI Agency chatbot
-- Date: 2025-12-02

CREATE TABLE IF NOT EXISTS aiconversationsessions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    session_id VARCHAR(36) NOT NULL UNIQUE,
    user_id VARCHAR(36) NOT NULL,
    editor_type VARCHAR(50) COMMENT 'Type of editor: bpm, page, view, workflow, whiteboard, report, general',
    context_data TEXT COMMENT 'JSON serialized conversation context',

    -- Standard IAC audit fields
    active BOOLEAN DEFAULT TRUE,
    referenceid VARCHAR(36),
    createdby VARCHAR(45),
    createdon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(45),
    modifiedon TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    rowversionstamp INT DEFAULT 1,

    INDEX idx_session_user (session_id, user_id),
    INDEX idx_user_active (user_id, active),
    INDEX idx_editor_type (editor_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='AI conversation sessions for unified AI agency chatbot';
