# Integration Tests Documentation

This document describes the integration tests implemented for the MAWB System Integration feature.

## Overview

The integration tests are divided into two main categories:

1. **Database Integration Tests** (`integration_test.go`) - Tests database operations, foreign key relationships, and data consistency
2. **API Integration Tests** (`api_integration_test.go`) - Tests HTTP endpoints, authentication flows, and complete request/response cycles

## Prerequisites

### Database Setup

Before running integration tests, you need to set up a test database:

1. Create a test PostgreSQL database
2. Set the following environment variables:

```bash
export TEST_POSTGRESQL_HOST=localhost
export TEST_POSTGRESQL_USER=postgres
export TEST_POSTGRESQL_PASSWORD=your_password
export TEST_POSTGRESQL_NAME=test_db
export TEST_POSTGRESQL_PORT=5432
export TEST_POSTGRESQL_SSLMODE=false
export TEST_INTEGRATION=1
```

### Dependencies

Ensure you have the required Go modules:

```bash
go mod tidy
```

## Running Integration Tests

### Run All Integration Tests

```bash
TEST_INTEGRATION=1 go test -v ./... -run "Integration"
```

### Run Database Integration Tests Only

```bash
TEST_INTEGRATION=1 go test -v -run "TestDatabaseIntegrationTestSuite"
```

### Run API Integration Tests Only

```bash
TEST_INTEGRATION=1 go test -v -run "TestAPIIntegrationTestSuite"
```

### Run Specific Test Cases

```bash
# Run specific database test
TEST_INTEGRATION=1 go test -v -run "TestDatabaseIntegrationTestSuite/TestCargoManifestCRUDOperations"

# Run specific API test
TEST_INTEGRATION=1 go test -v -run "TestAPIIntegrationTestSuite/TestCargoManifestAPIEndpoints"
```

## Database Integration Tests

### Test Coverage

The database integration tests cover:

#### CRUD Operations
- **TestCargoManifestCRUDOperations**: Complete Create, Read, Update, Delete operations for cargo manifests
- **TestDraftMAWBCRUDOperations**: Complete CRUD operations for draft MAWB with nested data structures

#### Foreign Key Relationships
- **TestForeignKeyRelationships**: Tests foreign key constraints and validation
- **TestCascadingDeletes**: Tests CASCADE DELETE operations across related tables

#### Transaction Management
- **TestTransactionRollback**: Tests transaction rollback scenarios
- **TestDataConsistency**: Tests data consistency across related tables

#### Performance and Concurrency
- **TestConcurrentAccess**: Tests concurrent database access patterns
- **TestDatabaseIndexes**: Tests database index performance

### Key Features Tested

1. **Foreign Key Constraints**: Validates that invalid MAWB Info UUIDs are rejected
2. **Cascading Deletes**: Ensures that deleting MAWB Info records cascades to related cargo manifest and draft MAWB data
3. **Data Integrity**: Tests complex nested data structures (items, dimensions, charges)
4. **Transaction Safety**: Validates rollback behavior on errors
5. **Concurrent Access**: Tests multiple simultaneous database operations

## API Integration Tests

### Test Coverage

The API integration tests cover:

#### Endpoint Testing
- **TestCargoManifestAPIEndpoints**: Tests all cargo manifest endpoints (GET, POST, confirm, reject, print)
- **TestDraftMAWBAPIEndpoints**: Tests all draft MAWB endpoints with complex nested data

#### Authentication and Authorization
- **TestAuthenticationFlows**: Tests JWT authentication scenarios
- Tests unauthorized access attempts
- Tests invalid token handling

#### Error Handling
- **TestErrorResponseFormats**: Tests error response formats and HTTP status codes
- Tests invalid UUID formats
- Tests missing required fields
- Tests malformed JSON payloads

#### PDF Generation
- **TestPDFGenerationEndpoints**: Tests PDF generation endpoints
- Validates PDF content type headers
- Basic PDF content validation (magic number check)

#### Advanced Scenarios
- **TestConcurrentAPIRequests**: Tests concurrent API requests
- **TestCompleteWorkflow**: Tests end-to-end workflow from creation to PDF generation
- **TestRequestValidation**: Comprehensive request validation testing

### Key Features Tested

1. **Full Request/Response Cycles**: Complete HTTP request/response testing
2. **Authentication Flows**: JWT token validation and error handling
3. **Data Validation**: Request payload validation and error responses
4. **Status Management**: Document status transitions (Draft → Confirmed → Rejected)
5. **PDF Generation**: File generation and content validation
6. **Error Scenarios**: Comprehensive error handling and response formatting

## Test Data Management

Both test suites implement comprehensive test data cleanup:

### Automatic Cleanup
- **SetupTest/TearDownTest**: Cleans data before/after each test
- **SetupSuite/TearDownSuite**: Cleans data before/after entire test suite

### Tracking
- Test UUIDs are tracked in slices for efficient cleanup
- Cleanup follows reverse dependency order to avoid foreign key violations

### Isolation
- Each test creates its own test data
- Tests are isolated from each other
- Database transactions can be used for additional isolation

## Test Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `TEST_POSTGRESQL_HOST` | Test database host | `localhost` |
| `TEST_POSTGRESQL_USER` | Test database user | `postgres` |
| `TEST_POSTGRESQL_PASSWORD` | Test database password | `password` |
| `TEST_POSTGRESQL_NAME` | Test database name | `test_db` |
| `TEST_POSTGRESQL_PORT` | Test database port | `5432` |
| `TEST_POSTGRESQL_SSLMODE` | SSL mode | `false` |
| `TEST_INTEGRATION` | Enable integration tests | (required) |

### Test Database Schema

The tests automatically create the required database schema:

1. **tbl_mawb_info**: Main MAWB Info table
2. **cargo_manifest**: Cargo manifest table with foreign key to MAWB Info
3. **cargo_manifest_items**: Cargo manifest items table
4. **draft_mawb**: Draft MAWB table with foreign key to MAWB Info
5. **draft_mawb_items**: Draft MAWB items table
6. **draft_mawb_item_dims**: Draft MAWB item dimensions table
7. **draft_mawb_charges**: Draft MAWB charges table

All tables include proper foreign key constraints with CASCADE DELETE.

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Verify database is running
   - Check connection parameters
   - Ensure test database exists

2. **Tests Skipped**
   - Ensure `TEST_INTEGRATION=1` is set
   - Check environment variable spelling

3. **Foreign Key Violations**
   - Tests should handle cleanup automatically
   - Check if previous test runs left orphaned data

4. **Authentication Failures**
   - API tests use mock JWT tokens
   - Ensure server authentication is properly configured for testing

### Debug Mode

Run tests with verbose output:

```bash
TEST_INTEGRATION=1 go test -v -run "Integration" -count=1
```

Add `-count=1` to disable test caching for debugging.

## Performance Considerations

### Database Tests
- Tests include performance validation for indexed queries
- Concurrent access tests validate thread safety
- Transaction tests ensure proper rollback behavior

### API Tests
- Concurrent request testing validates server performance
- PDF generation tests include timeout handling
- Large payload tests validate request size limits

## Requirements Coverage

These integration tests fulfill the following requirements from the specification:

### Requirement 8.2 (Integration Tests)
- ✅ Database operations with real database connections
- ✅ Foreign key relationships and cascading operations
- ✅ Transaction rollback scenarios
- ✅ Data consistency validation

### Requirement 8.4 (API Testing)
- ✅ Full request/response cycle tests
- ✅ Authentication flows and authorization scenarios
- ✅ PDF generation endpoints with file validation
- ✅ Error response formats and status codes

### Requirements 6.1-6.5 (Database Schema)
- ✅ Foreign key constraint validation
- ✅ CASCADE DELETE operations
- ✅ Data integrity across related tables
- ✅ Index performance validation

### Requirements 4.1-4.4 (PDF Generation)
- ✅ PDF generation endpoint testing
- ✅ Content-type header validation
- ✅ Basic PDF content validation
- ✅ Error handling for PDF generation failures

## Continuous Integration

These tests can be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions configuration
- name: Run Integration Tests
  env:
    TEST_INTEGRATION: 1
    TEST_POSTGRESQL_HOST: localhost
    TEST_POSTGRESQL_USER: postgres
    TEST_POSTGRESQL_PASSWORD: postgres
    TEST_POSTGRESQL_NAME: test_db
  run: |
    go test -v ./... -run "Integration"
```

## Maintenance

### Adding New Tests
1. Follow the existing test suite patterns
2. Ensure proper test data cleanup
3. Add appropriate assertions for new functionality
4. Update this documentation

### Updating Schema
1. Update migration functions in test files
2. Ensure backward compatibility
3. Test with both empty and populated databases

### Performance Monitoring
1. Monitor test execution times
2. Add benchmarks for critical operations
3. Validate index effectiveness with larger datasets