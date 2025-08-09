# Database Migrations

This directory contains SQL migration files for the MAWB System Integration feature.

## Migration Files

- `001_create_cargo_manifest_tables.sql` - Creates cargo_manifest and cargo_manifest_items tables
- `002_create_draft_mawb_tables.sql` - Creates draft_mawb, draft_mawb_items, draft_mawb_item_dims, and draft_mawb_charges tables
- `003_rollback_mawb_system_integration.sql` - Rollback migration (for development/testing)

## Running Migrations

### Using Make (Recommended)
```bash
make migrate
```

### Using Go directly
```bash
go run cmd/migrate/main.go
```

## Database Schema

### Tables Created

#### cargo_manifest
- Primary table for cargo manifest data
- Foreign key to `tbl_mawb_info(uuid)` with CASCADE DELETE
- Status field with check constraint ('Draft', 'Pending', 'Confirmed', 'Rejected')

#### cargo_manifest_items
- Items belonging to a cargo manifest
- Foreign key to `cargo_manifest(uuid)` with CASCADE DELETE

#### draft_mawb
- Primary table for draft MAWB data
- Foreign key to `tbl_mawb_info(uuid)` with CASCADE DELETE
- Contains 40+ fields for comprehensive MAWB information
- Status field with check constraint ('Draft', 'Pending', 'Confirmed', 'Rejected')

#### draft_mawb_items
- Items belonging to a draft MAWB
- Foreign key to `draft_mawb(uuid)` with CASCADE DELETE

#### draft_mawb_item_dims
- Dimension data for draft MAWB items
- Foreign key to `draft_mawb_items(id)` with CASCADE DELETE

#### draft_mawb_charges
- Charge data for draft MAWB
- Foreign key to `draft_mawb(uuid)` with CASCADE DELETE
- Key-value structure for flexible charge types

### Indexes Created

Performance indexes are created on:
- All foreign key columns
- Status fields for filtering
- Created_at fields for date-based queries

### Triggers

- `update_updated_at_column()` function and triggers for automatic timestamp updates

## Foreign Key Relationships

```
tbl_mawb_info (existing)
├── cargo_manifest (CASCADE DELETE)
│   └── cargo_manifest_items (CASCADE DELETE)
└── draft_mawb (CASCADE DELETE)
    ├── draft_mawb_items (CASCADE DELETE)
    │   └── draft_mawb_item_dims (CASCADE DELETE)
    └── draft_mawb_charges (CASCADE DELETE)
```

## Requirements Satisfied

This migration satisfies the following requirements from the specification:

- **6.1**: CASCADE DELETE relationships when MAWB Info records are deleted
- **6.2**: CASCADE DELETE for cargo manifest and related items
- **6.3**: CASCADE DELETE for draft MAWB and related items, dimensions, and charges
- **6.4**: CASCADE DELETE for draft MAWB items and related dimensions
- **6.5**: Foreign key constraint validation and error handling

## Notes

- All tables use UUID primary keys (except item tables which use SERIAL)
- Timestamps are automatically managed with triggers
- Status fields have check constraints to ensure data integrity
- Indexes are optimized for the expected query patterns
- CASCADE DELETE ensures data consistency across related tables