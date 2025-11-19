-- Migration: Add support for page-break, page-header, and page-footer component types
-- Date: 2025-11-19

-- Add new columns for page component configurations
ALTER TABLE reportcomponents
ADD COLUMN pagebreakconfig JSON AFTER drilldownconfig,
ADD COLUMN pageheaderconfig JSON AFTER pagebreakconfig,
ADD COLUMN pagefooterconfig JSON AFTER pageheaderconfig;

-- Update the componenttype ENUM to include new types
ALTER TABLE reportcomponents
MODIFY COLUMN componenttype ENUM(
  'table',
  'chart',
  'barcode',
  'sub_report',
  'text',
  'image',
  'drill_down',
  'page-break',
  'page-header',
  'page-footer'
) NOT NULL;
