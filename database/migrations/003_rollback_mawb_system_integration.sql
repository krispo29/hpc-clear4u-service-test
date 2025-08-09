-- Migration: Rollback MAWB System Integration
-- Description: Drops all tables created for MAWB system integration (for development/testing purposes)
-- WARNING: This will permanently delete all data in these tables

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS draft_mawb_charges CASCADE;
DROP TABLE IF EXISTS draft_mawb_item_dims CASCADE;
DROP TABLE IF EXISTS draft_mawb_items CASCADE;
DROP TABLE IF EXISTS draft_mawb CASCADE;
DROP TABLE IF EXISTS cargo_manifest_items CASCADE;
DROP TABLE IF EXISTS cargo_manifest CASCADE;

-- Drop the update trigger function if no other tables are using it
-- (Commented out to be safe - only uncomment if you're sure no other tables use this function)
-- DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;