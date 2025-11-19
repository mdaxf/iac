-- Migration: Add support for page-break, page-header, and page-footer component types
-- Date: 2025-11-19

-- Add new columns for page component configurations
ALTER TABLE reportcomponents
ADD COLUMN pagebreakconfig JSONB,
ADD COLUMN pageheaderconfig JSONB,
ADD COLUMN pagefooterconfig JSONB;

-- Update the componenttype type to include new types
-- First, add the new values to the enum type
ALTER TYPE componenttype ADD VALUE IF NOT EXISTS 'page-break';
ALTER TYPE componenttype ADD VALUE IF NOT EXISTS 'page-header';
ALTER TYPE componenttype ADD VALUE IF NOT EXISTS 'page-footer';
