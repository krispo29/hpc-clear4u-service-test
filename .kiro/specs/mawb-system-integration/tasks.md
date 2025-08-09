# Implementation Plan

- [ ] 1. Set up database schema and migrations
  - Create database migration files for all new tables (cargo_manifest, cargo_manifest_items, draft_mawb, draft_mawb_items, draft_mawb_item_dims, draft_mawb_charges)
  - Implement foreign key constraints with CASCADE DELETE relationships
  - Add database indexes for performance optimization on foreign keys and status fields
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [ ] 2. Create Cargo Manifest data models and validation
  - [ ] 2.1 Implement Cargo Manifest core data structures
    - Write Go structs for CargoManifest, CargoManifestItem with proper JSON and database tags
    - Create request/response models for CargoManifestRequest, CargoManifestItemRequest
    - Implement validation tags and custom validation functions
    - _Requirements: 1.1, 1.2, 5.3_

  - [ ] 2.2 Create Cargo Manifest repository interface and implementation
    - Define CargoManifestRepository interface with CRUD operations
    - Implement repository with PostgreSQL operations using go-pg patterns
    - Add methods for GetByMAWBUUID, CreateOrUpdate, UpdateStatus, ValidateMAWBExists
    - Implement transaction handling for multi-table operations
    - _Requirements: 1.1, 1.2, 1.4, 5.1, 6.1_

  - [ ] 2.3 Implement Cargo Manifest service layer
    - Create CargoManifestService interface with business logic methods
    - Implement service with validation, error handling, and business rules
    - Add methods for GetCargoManifest, CreateOrUpdateCargoManifest, status management
    - Implement MAWB Info existence validation
    - _Requirements: 1.1, 1.2, 1.3, 1.5, 5.1, 5.2_

- [ ] 3. Create Draft MAWB data models and validation
  - [ ] 3.1 Implement Draft MAWB core data structures
    - Write Go structs for DraftMAWB, DraftMAWBItem, DraftMAWBItemDim, DraftMAWBCharge
    - Create comprehensive request/response models with all 40+ fields
    - Implement validation tags and business rule validation functions
    - _Requirements: 2.1, 2.2, 2.3, 5.3_

  - [ ] 3.2 Create Draft MAWB repository interface and implementation
    - Define DraftMAWBRepository interface with complex CRUD operations
    - Implement repository with nested data handling (items, dimensions, charges)
    - Add transaction management for multi-table operations with proper rollback
    - Implement efficient JOIN queries for nested data retrieval
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 6.1, 6.3, 6.4_

  - [ ] 3.3 Implement Draft MAWB service layer with calculations
    - Create DraftMAWBService interface with calculation methods
    - Implement volumetric weight calculation using (L×W×H)/1,000,000 × count formula
    - Add chargeable weight calculation (max of actual weight and volumetric weight × 166.67)
    - Implement financial totals calculation for charges, taxes, and fees
    - Add unit conversion logic for pounds to kilograms (× 0.453592)
    - _Requirements: 2.1, 2.2, 2.4, 2.5, 7.1, 7.2, 7.3, 7.4, 7.5_

- [ ] 4. Implement HTTP handlers and routing
  - [ ] 4.1 Create Cargo Manifest HTTP handlers
    - Implement cargoManifestHandler struct with service dependency
    - Create handlers for GET, POST, confirm, reject, and print endpoints
    - Add request binding, validation, and error response handling
    - Implement proper HTTP status codes and response formatting
    - _Requirements: 1.1, 1.2, 3.1, 3.2, 4.1, 5.1, 5.4_

  - [ ] 4.2 Create Draft MAWB HTTP handlers
    - Implement draftMAWBHandler struct with service dependency
    - Create handlers for GET, POST, confirm, reject, and print endpoints
    - Add complex request binding for nested data structures
    - Implement comprehensive validation and error handling
    - _Requirements: 2.1, 2.2, 3.3, 3.4, 4.2, 5.1, 5.4_

  - [ ] 4.3 Extend MAWB Info handler with sub-routes
    - Modify existing mawbInfoHandler to include cargo manifest and draft MAWB sub-routes
    - Implement Chi router mounting for /cargo-manifest and /draft-mawb paths
    - Add middleware for UUID parameter validation
    - Ensure proper authentication flow through sub-routes
    - _Requirements: 1.1, 2.1, 5.1, 5.4_

- [ ] 5. Integrate with factory pattern and dependency injection
  - [ ] 5.1 Update repository factory
    - Add CargoManifestRepository and DraftMAWBRepository to RepositoryFactory struct
    - Implement factory methods with proper timeout configuration
    - Update NewRepositoryFactory to instantiate new repositories
    - _Requirements: 1.1, 2.1_

  - [ ] 5.2 Update service factory
    - Add CargoManifestService and DraftMAWBService to ServiceFactory struct
    - Implement factory methods with repository dependencies and timeout configuration
    - Update NewServiceFactory to instantiate new services with proper dependencies
    - _Requirements: 1.1, 2.1_

  - [ ] 5.3 Update server routing integration
    - Modify server.go to include new services in handler initialization
    - Update mawbInfoHandler constructor to accept new service dependencies
    - Ensure proper dependency injection chain from main.go through factories
    - _Requirements: 1.1, 2.1_

- [ ] 6. Implement PDF generation functionality
  - [ ] 6.1 Create PDF generation service
    - Implement PDFGenerator struct with template and font management
    - Create GenerateCargoManifestPDF method using existing font loading patterns
    - Implement GenerateDraftMAWBPDF method with complex layout handling
    - Add error handling and fallback mechanisms for PDF generation failures
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

  - [ ] 6.2 Create PDF templates
    - Design cargo manifest PDF template with proper formatting and branding
    - Create draft MAWB PDF template handling all fields and nested data
    - Implement multi-language support following existing patterns
    - Add template caching for performance optimization
    - _Requirements: 4.1, 4.2, 4.3_

  - [ ] 6.3 Integrate PDF endpoints
    - Add PDF generation to cargo manifest and draft MAWB handlers
    - Implement proper content-type headers and file download responses
    - Add error handling for PDF generation failures with appropriate HTTP responses
    - _Requirements: 4.1, 4.2, 4.4_

- [ ] 7. Implement comprehensive error handling
  - [ ] 7.1 Create custom error types
    - Define MAWBNotFoundError, CargoManifestNotFoundError, DraftMAWBNotFoundError
    - Implement ValidationError for business rule violations
    - Create error mapping functions for consistent HTTP responses
    - _Requirements: 5.1, 5.2, 5.4, 5.5_

  - [ ] 7.2 Add error handling middleware
    - Implement error recovery middleware for panic situations
    - Add request logging with error context for debugging
    - Create error response standardization across all endpoints
    - _Requirements: 5.4, 5.5_

- [ ] 8. Write comprehensive unit tests
  - [ ] 8.1 Create service layer unit tests
    - Write tests for CargoManifestService with mocked repository dependencies
    - Implement DraftMAWBService tests including calculation logic validation
    - Add test cases for error scenarios and edge cases
    - Achieve minimum 80% code coverage for service layer
    - _Requirements: 8.1, 8.4_

  - [ ] 8.2 Create repository layer unit tests
    - Write tests for CargoManifestRepository with test database
    - Implement DraftMAWBRepository tests with transaction validation
    - Add tests for foreign key constraint validation and cascading deletes
    - Test concurrent access patterns and data integrity
    - _Requirements: 8.1, 8.4, 6.1, 6.2, 6.3_

  - [ ] 8.3 Create handler layer unit tests
    - Write tests for cargoManifestHandler with mocked service dependencies
    - Implement draftMAWBHandler tests with complex request/response validation
    - Add tests for authentication middleware integration
    - Test error response formatting and HTTP status codes
    - _Requirements: 8.1, 8.4, 5.4_

- [ ] 9. Write integration tests
  - [ ] 9.1 Create database integration tests
    - Write end-to-end tests for complete CRUD operations
    - Test foreign key relationships and cascading operations
    - Add tests for transaction rollback scenarios
    - Validate data consistency across related tables
    - _Requirements: 8.2, 8.4, 6.1, 6.2, 6.3, 6.4_

  - [ ] 9.2 Create API integration tests
    - Write full request/response cycle tests for all endpoints
    - Test authentication flows and authorization scenarios
    - Add tests for PDF generation endpoints with actual file validation
    - Validate error response formats and status codes
    - _Requirements: 8.2, 8.4, 4.1, 4.2, 4.3, 4.4_

- [ ] 10. Implement performance optimizations
  - [ ] 10.1 Add database indexing and query optimization
    - Create database indexes for foreign keys and frequently queried fields
    - Optimize JOIN queries for nested data retrieval
    - Implement query performance monitoring and logging
    - _Requirements: 8.5_

  - [ ] 10.2 Implement caching strategy
    - Add caching for frequently accessed MAWB Info records
    - Implement PDF template caching for generation performance
    - Create calculation result caching for complex computations
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_

- [ ] 11. Add comprehensive validation and security
  - [ ] 11.1 Implement input validation and sanitization
    - Add comprehensive input validation for all request models
    - Implement business rule validation for cargo manifest and draft MAWB data
    - Create input sanitization to prevent injection attacks
    - Add request size limits and rate limiting
    - _Requirements: 5.1, 5.2, 5.3, 5.5_

  - [ ] 11.2 Add security enhancements
    - Implement audit logging for sensitive operations
    - Add role-based access control for status change operations
    - Create data masking for sensitive information in logs
    - Implement secure PDF generation with proper error handling
    - _Requirements: 5.1, 5.4, 5.5_

- [ ] 12. Final integration and testing
  - [ ] 12.1 Integration testing with existing MAWB Info system
    - Test complete workflow from MAWB Info creation to cargo manifest and draft MAWB
    - Validate foreign key relationships work correctly with existing data
    - Test cascading delete operations don't affect unrelated data
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

  - [ ] 12.2 End-to-end system testing
    - Perform complete system testing with realistic data volumes
    - Test concurrent user scenarios and data consistency
    - Validate PDF generation under load conditions
    - Run performance tests and optimize bottlenecks
    - _Requirements: 8.5, 7.1, 7.2, 7.3, 7.4, 7.5_