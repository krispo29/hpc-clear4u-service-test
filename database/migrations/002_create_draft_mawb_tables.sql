-- Migration: Create draft MAWB tables
-- Description: Creates draft_mawb, draft_mawb_items, draft_mawb_item_dims, and draft_mawb_charges tables

-- Create draft_mawb table
CREATE TABLE IF NOT EXISTS draft_mawb (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mawb_info_uuid UUID NOT NULL,
    customer_uuid UUID,
    airline_logo VARCHAR(255),
    airline_name VARCHAR(255),
    mawb VARCHAR(255) NOT NULL,
    hawb VARCHAR(255),
    shipper_name_and_address TEXT,
    consignee_name_and_address TEXT,
    issuing_carrier_agent_name_and_city VARCHAR(255),
    accounting_information TEXT,
    agent_iata_code VARCHAR(50),
    account_no VARCHAR(100),
    airport_of_departure VARCHAR(100),
    reference_number VARCHAR(100),
    to_1 VARCHAR(100),
    by_first_carrier VARCHAR(100),
    to_2 VARCHAR(100),
    by_2 VARCHAR(100),
    to_3 VARCHAR(100),
    by_3 VARCHAR(100),
    currency VARCHAR(10),
    chgs_code VARCHAR(10),
    wt_val_ppd VARCHAR(50),
    wt_val_coll VARCHAR(50),
    other_ppd VARCHAR(50),
    other_coll VARCHAR(50),
    declared_value_carriage VARCHAR(100),
    declared_value_customs VARCHAR(100),
    airport_of_destination VARCHAR(100),
    flight_no VARCHAR(100),
    flight_date DATE,
    insurance_amount DECIMAL(15,2),
    handling_information TEXT,
    sci VARCHAR(255),
    total_no_of_pieces INTEGER,
    total_gross_weight DECIMAL(10,2),
    total_kg_lb VARCHAR(10),
    total_rate_class VARCHAR(50),
    total_chargeable_weight DECIMAL(10,2),
    total_rate_charge DECIMAL(15,2),
    total_amount DECIMAL(15,2),
    shipper_certifies_text TEXT,
    executed_on_date DATE,
    executed_at_place VARCHAR(255),
    signature_of_shipper VARCHAR(255),
    signature_of_issuing_carrier VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'Draft',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraint with CASCADE DELETE
    CONSTRAINT fk_draft_mawb_mawb_info 
        FOREIGN KEY (mawb_info_uuid) 
        REFERENCES tbl_mawb_info(uuid) 
        ON DELETE CASCADE,
    
    -- Check constraint for status values
    CONSTRAINT chk_draft_mawb_status 
        CHECK (status IN ('Draft', 'Pending', 'Confirmed', 'Rejected'))
);

-- Create draft_mawb_items table
CREATE TABLE IF NOT EXISTS draft_mawb_items (
    id SERIAL PRIMARY KEY,
    draft_mawb_uuid UUID NOT NULL,
    pieces_rcp VARCHAR(100),
    gross_weight VARCHAR(100),
    kg_lb VARCHAR(10),
    rate_class VARCHAR(50),
    total_volume DECIMAL(15,6),
    chargeable_weight DECIMAL(10,2),
    rate_charge DECIMAL(15,2),
    total DECIMAL(15,2),
    nature_and_quantity TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraint with CASCADE DELETE
    CONSTRAINT fk_draft_mawb_items_draft_mawb 
        FOREIGN KEY (draft_mawb_uuid) 
        REFERENCES draft_mawb(uuid) 
        ON DELETE CASCADE
);

-- Create draft_mawb_item_dims table
CREATE TABLE IF NOT EXISTS draft_mawb_item_dims (
    id SERIAL PRIMARY KEY,
    draft_mawb_item_id INTEGER NOT NULL,
    length VARCHAR(50),
    width VARCHAR(50),
    height VARCHAR(50),
    count VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraint with CASCADE DELETE
    CONSTRAINT fk_draft_mawb_item_dims_draft_mawb_item 
        FOREIGN KEY (draft_mawb_item_id) 
        REFERENCES draft_mawb_items(id) 
        ON DELETE CASCADE
);

-- Create draft_mawb_charges table
CREATE TABLE IF NOT EXISTS draft_mawb_charges (
    id SERIAL PRIMARY KEY,
    draft_mawb_uuid UUID NOT NULL,
    charge_key VARCHAR(100) NOT NULL,
    charge_value DECIMAL(15,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraint with CASCADE DELETE
    CONSTRAINT fk_draft_mawb_charges_draft_mawb 
        FOREIGN KEY (draft_mawb_uuid) 
        REFERENCES draft_mawb(uuid) 
        ON DELETE CASCADE
);

-- Create indexes for performance optimization
CREATE INDEX IF NOT EXISTS idx_draft_mawb_mawb_info_uuid 
    ON draft_mawb(mawb_info_uuid);

CREATE INDEX IF NOT EXISTS idx_draft_mawb_status 
    ON draft_mawb(status);

CREATE INDEX IF NOT EXISTS idx_draft_mawb_created_at 
    ON draft_mawb(created_at);

CREATE INDEX IF NOT EXISTS idx_draft_mawb_items_draft_mawb_uuid 
    ON draft_mawb_items(draft_mawb_uuid);

CREATE INDEX IF NOT EXISTS idx_draft_mawb_item_dims_draft_mawb_item_id 
    ON draft_mawb_item_dims(draft_mawb_item_id);

CREATE INDEX IF NOT EXISTS idx_draft_mawb_charges_draft_mawb_uuid 
    ON draft_mawb_charges(draft_mawb_uuid);

-- Add updated_at trigger for draft_mawb
CREATE TRIGGER update_draft_mawb_updated_at 
    BEFORE UPDATE ON draft_mawb 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();