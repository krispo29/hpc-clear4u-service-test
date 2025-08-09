-- Migration: Add performance optimization indexes
-- Description: Creates additional indexes for query optimization and performance monitoring

-- Additional composite indexes for complex queries
CREATE INDEX IF NOT EXISTS idx_cargo_manifest_status_created_at 
    ON cargo_manifest(status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_cargo_manifest_mawb_number_status 
    ON cargo_manifest(mawb_number, status);

CREATE INDEX IF NOT EXISTS idx_draft_mawb_status_created_at 
    ON draft_mawb(status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_draft_mawb_customer_status 
    ON draft_mawb(customer_uuid, status) WHERE customer_uuid IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_draft_mawb_mawb_status 
    ON draft_mawb(mawb, status);

-- Indexes for frequently searched text fields
CREATE INDEX IF NOT EXISTS idx_cargo_manifest_items_hawb_no 
    ON cargo_manifest_items(hawb_no) WHERE hawb_no IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_cargo_manifest_items_destination 
    ON cargo_manifest_items(destination) WHERE destination IS NOT NULL;

-- Indexes for draft MAWB items calculations
CREATE INDEX IF NOT EXISTS idx_draft_mawb_items_chargeable_weight 
    ON draft_mawb_items(chargeable_weight) WHERE chargeable_weight IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_draft_mawb_items_rate_class 
    ON draft_mawb_items(rate_class) WHERE rate_class IS NOT NULL;

-- Indexes for charges lookup
CREATE INDEX IF NOT EXISTS idx_draft_mawb_charges_key 
    ON draft_mawb_charges(charge_key);

CREATE INDEX IF NOT EXISTS idx_draft_mawb_charges_key_value 
    ON draft_mawb_charges(charge_key, charge_value);

-- Partial indexes for active records (non-deleted status)
CREATE INDEX IF NOT EXISTS idx_cargo_manifest_active 
    ON cargo_manifest(mawb_info_uuid, updated_at DESC) 
    WHERE status != 'Rejected';

CREATE INDEX IF NOT EXISTS idx_draft_mawb_active 
    ON draft_mawb(mawb_info_uuid, updated_at DESC) 
    WHERE status != 'Rejected';

-- Indexes for date range queries
CREATE INDEX IF NOT EXISTS idx_cargo_manifest_freight_date 
    ON cargo_manifest(freight_date) WHERE freight_date IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_draft_mawb_flight_date 
    ON draft_mawb(flight_date) WHERE flight_date IS NOT NULL;

-- Covering indexes for common SELECT queries
CREATE INDEX IF NOT EXISTS idx_cargo_manifest_list_covering 
    ON cargo_manifest(mawb_info_uuid, status, created_at DESC) 
    INCLUDE (uuid, mawb_number, shipper, consignee);

CREATE INDEX IF NOT EXISTS idx_draft_mawb_list_covering 
    ON draft_mawb(mawb_info_uuid, status, created_at DESC) 
    INCLUDE (uuid, mawb, customer_uuid, airline_name);

-- Function-based index for case-insensitive MAWB number searches
CREATE INDEX IF NOT EXISTS idx_cargo_manifest_mawb_number_lower 
    ON cargo_manifest(LOWER(mawb_number));

CREATE INDEX IF NOT EXISTS idx_draft_mawb_mawb_lower 
    ON draft_mawb(LOWER(mawb));

-- Statistics update for query planner optimization
ANALYZE cargo_manifest;
ANALYZE cargo_manifest_items;
ANALYZE draft_mawb;
ANALYZE draft_mawb_items;
ANALYZE draft_mawb_item_dims;
ANALYZE draft_mawb_charges;