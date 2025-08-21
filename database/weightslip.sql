CREATE TABLE IF NOT EXISTS public.weight_slip (
    uuid UUID PRIMARY KEY,
    mawb_info_uuid UUID REFERENCES mawb_info(uuid),
    slip_no VARCHAR(255),
    wsid VARCHAR(255),
    date_time TIMESTAMPTZ,
    pseq VARCHAR(255),
    staff VARCHAR(255),
    mawb VARCHAR(255),
    hawb VARCHAR(255),
    dest VARCHAR(255),
    agent_code VARCHAR(255),
    agent_name VARCHAR(255),
    flight VARCHAR(255),
    nature_of_goods VARCHAR(255),
    ews BOOLEAN,
    pcs INT,
    gw NUMERIC,
    tw NUMERIC,
    nw NUMERIC,
    dim_weight NUMERIC,
    volume_m3 NUMERIC,
    status_uuid UUID REFERENCES master_status(uuid),
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS public.weight_slip_dimensions (
    id SERIAL PRIMARY KEY,
    weightslip_uuid UUID REFERENCES weight_slip(uuid) ON DELETE CASCADE,
    no INT,
    l_cm NUMERIC,
    w_cm NUMERIC,
    h_cm NUMERIC,
    pcs INT
);
