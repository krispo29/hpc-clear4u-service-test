-- Migration: Create cargo manifest tables
-- Description: Creates cargo_manifest and cargo_manifest_items tables with foreign key relationships

-- Create cargo_manifest table
CREATE TABLE IF NOT EXISTS cargo_manifest (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mawb_info_uuid UUID NOT NULL,
    mawb_number VARCHAR(255) NOT NULL,
    port_of_discharge VARCHAR(255),
    flight_no VARCHAR(100),
    freight_date DATE,
    shipper TEXT,
    consignee TEXT,
    total_ctn VARCHAR(50),
    transshipment VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'Draft',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraint with CASCADE DELETE
    CONSTRAINT fk_cargo_manifest_mawb_info 
        FOREIGN KEY (mawb_info_uuid) 
        REFERENCES tbl_mawb_info(uuid) 
        ON DELETE CASCADE,
    
    -- Check constraint for status values
    CONSTRAINT chk_cargo_manifest_status 
        CHECK (status IN ('Draft', 'Pending', 'Confirmed', 'Rejected'))
);

-- Create cargo_manifest_items table
CREATE TABLE IF NOT EXISTS cargo_manifest_items (
    id SERIAL PRIMARY KEY,
    cargo_manifest_uuid UUID NOT NULL,
    hawb_no VARCHAR(255),
    pkgs VARCHAR(100),
    gross_weight VARCHAR(100),
    destination VARCHAR(255),
    commodity TEXT,
    shipper_name_address TEXT,
    consignee_name_address TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraint with CASCADE DELETE
    CONSTRAINT fk_cargo_manifest_items_cargo_manifest 
        FOREIGN KEY (cargo_manifest_uuid) 
        REFERENCES cargo_manifest(uuid) 
        ON DELETE CASCADE
);

-- Create indexes for performance optimization
CREATE INDEX IF NOT EXISTS idx_cargo_manifest_mawb_info_uuid 
    ON cargo_manifest(mawb_info_uuid);

CREATE INDEX IF NOT EXISTS idx_cargo_manifest_status 
    ON cargo_manifest(status);

CREATE INDEX IF NOT EXISTS idx_cargo_manifest_created_at 
    ON cargo_manifest(created_at);

CREATE INDEX IF NOT EXISTS idx_cargo_manifest_items_cargo_manifest_uuid 
    ON cargo_manifest_items(cargo_manifest_uuid);

-- Add updated_at trigger for cargo_manifest
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_cargo_manifest_updated_at 
    BEFORE UPDATE ON cargo_manifest 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();