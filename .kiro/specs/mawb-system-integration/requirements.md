# Requirements Document

## Introduction

This feature involves developing comprehensive API endpoints to integrate MAWB Info with Cargo Manifest and Draft MAWB systems using foreign key relationships. The system will enable users to create, retrieve, update, and manage cargo manifests and draft MAWB documents that are linked to existing MAWB Info records. The integration includes complex data structures with nested items, dimensions, and charges, along with status management and PDF generation capabilities.

## Requirements

### Requirement 1

**User Story:** As a logistics operator, I want to create and manage cargo manifests linked to MAWB Info records, so that I can track cargo details and maintain proper documentation relationships.

#### Acceptance Criteria

1. WHEN a user sends a POST request to `/v1/mawbinfo/{uuid}/cargo-manifest` with valid cargo manifest data THEN the system SHALL create a new cargo manifest record linked to the specified MAWB Info UUID
2. WHEN a user sends a GET request to `/v1/mawbinfo/{uuid}/cargo-manifest` THEN the system SHALL return the cargo manifest data with all associated items
3. WHEN a cargo manifest already exists for a MAWB Info UUID AND a user sends a POST request THEN the system SHALL update the existing cargo manifest record
4. WHEN a user provides cargo manifest items in the request THEN the system SHALL store each item with proper foreign key relationships
5. IF the specified MAWB Info UUID does not exist THEN the system SHALL return a 404 error with appropriate message

### Requirement 2

**User Story:** As a logistics operator, I want to create and manage draft MAWB documents with complex nested data structures, so that I can prepare detailed shipping documentation before finalization.

#### Acceptance Criteria

1. WHEN a user sends a POST request to `/v1/mawbinfo/{uuid}/draft-mawb` with valid draft MAWB data THEN the system SHALL create a new draft MAWB record with all nested items, dimensions, and charges
2. WHEN a user sends a GET request to `/v1/mawbinfo/{uuid}/draft-mawb` THEN the system SHALL return the complete draft MAWB structure including items, dimensions, and charges
3. WHEN draft MAWB items include dimension data THEN the system SHALL store each dimension record linked to the appropriate item
4. WHEN draft MAWB includes charges data THEN the system SHALL store each charge record with proper key-value mapping
5. WHEN calculating totals for draft MAWB items THEN the system SHALL compute volume, chargeable weight, and financial totals accurately

### Requirement 3

**User Story:** As a logistics supervisor, I want to manage the status of cargo manifests and draft MAWB documents, so that I can control the approval workflow and document lifecycle.

#### Acceptance Criteria

1. WHEN a user sends a POST request to `/v1/mawbinfo/{uuid}/cargo-manifest/confirm` THEN the system SHALL update the cargo manifest status to 'Confirmed'
2. WHEN a user sends a POST request to `/v1/mawbinfo/{uuid}/cargo-manifest/reject` THEN the system SHALL update the cargo manifest status to 'Rejected'
3. WHEN a user sends a POST request to `/v1/mawbinfo/{uuid}/draft-mawb/confirm` THEN the system SHALL update the draft MAWB status to 'Confirmed'
4. WHEN a user sends a POST request to `/v1/mawbinfo/{uuid}/draft-mawb/reject` THEN the system SHALL update the draft MAWB status to 'Rejected'
5. WHEN status updates occur THEN the system SHALL update the `updated_at` timestamp automatically

### Requirement 4

**User Story:** As a logistics operator, I want to generate PDF documents from cargo manifests and draft MAWB records, so that I can provide printed documentation for shipping and customs purposes.

#### Acceptance Criteria

1. WHEN a user sends a GET request to `/v1/mawbinfo/{uuid}/cargo-manifest/print` THEN the system SHALL generate and return a PDF document containing the cargo manifest data
2. WHEN a user sends a GET request to `/v1/mawbinfo/{uuid}/draft-mawb/print` THEN the system SHALL generate and return a PDF document containing the draft MAWB data
3. WHEN generating PDFs THEN the system SHALL include all relevant data fields in a properly formatted layout
4. WHEN PDF generation fails THEN the system SHALL return appropriate error responses
5. WHEN PDFs are generated THEN the system SHALL set proper content-type headers for file download

### Requirement 5

**User Story:** As a system administrator, I want robust data validation and error handling across all API endpoints, so that data integrity is maintained and users receive clear feedback on issues.

#### Acceptance Criteria

1. WHEN any API request is made with an invalid MAWB Info UUID THEN the system SHALL return a 404 error with message "MAWB Info not found"
2. WHEN required fields are missing from request bodies THEN the system SHALL return a 400 error with specific field validation messages
3. WHEN status values are provided THEN the system SHALL validate against allowed enum values ('Draft', 'Pending', 'Confirmed', 'Rejected')
4. WHEN database operations fail THEN the system SHALL return 500 errors with appropriate logging
5. WHEN authentication or authorization fails THEN the system SHALL return 401/403 errors respectively

### Requirement 6

**User Story:** As a system developer, I want proper database schema implementation with foreign key relationships and cascading deletes, so that data consistency is maintained across related tables.

#### Acceptance Criteria

1. WHEN MAWB Info records are deleted THEN the system SHALL cascade delete all related cargo manifest and draft MAWB records
2. WHEN cargo manifest records are deleted THEN the system SHALL cascade delete all related cargo manifest items
3. WHEN draft MAWB records are deleted THEN the system SHALL cascade delete all related items, dimensions, and charges
4. WHEN draft MAWB items are deleted THEN the system SHALL cascade delete all related dimension records
5. WHEN foreign key constraints are violated THEN the system SHALL prevent the operation and return appropriate errors

### Requirement 7

**User Story:** As a logistics operator, I want accurate calculations for volumetric weight, chargeable weight, and financial totals in draft MAWB documents, so that shipping costs and documentation are precise.

#### Acceptance Criteria

1. WHEN draft MAWB items include dimension data THEN the system SHALL calculate total volume using the formula (L×W×H)/1,000,000 × count
2. WHEN calculating chargeable weight THEN the system SHALL use the maximum of actual weight and volumetric weight (volume × 166.67)
3. WHEN weight units are in pounds THEN the system SHALL convert to kilograms using factor 0.453592
4. WHEN calculating financial totals THEN the system SHALL sum all charges, taxes, and fees accurately
5. WHEN item rate charges are provided THEN the system SHALL calculate item totals as rate × chargeable weight

### Requirement 8

**User Story:** As a system integrator, I want comprehensive API testing coverage including unit tests, integration tests, and error scenario testing, so that the system reliability is ensured.

#### Acceptance Criteria

1. WHEN unit tests are executed THEN the system SHALL achieve minimum 80% code coverage for all handler functions
2. WHEN integration tests are run THEN the system SHALL validate database operations with real database connections
3. WHEN API endpoint tests are executed THEN the system SHALL verify all request/response formats and status codes
4. WHEN error scenario tests are run THEN the system SHALL validate proper error handling for all failure cases
5. WHEN performance tests are conducted THEN the system SHALL handle concurrent requests without data corruption