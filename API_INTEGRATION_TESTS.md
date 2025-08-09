# API Integration Tests

This document describes the comprehensive API integration tests implemented for the MAWB System Integration feature.

## Overview

The API integration tests provide comprehensive coverage for all endpoints, authentication flows, error scenarios, and PDF generation functionality. These tests fulfill the requirements for task 9.2 "Create API integration tests".

## Test Coverage

### 1. Full Request/Response Cycle Tests

- **TestCargoManifestAPIEndpoints**: Tests all cargo manifest endpoints (GET, POST, confirm, reject, print)
- **TestDraftMAWBAPIEndpoints**: Tests all draft MAWB endpoints with complex nested data structures
- Complete CRUD operations testing
- Status transition workflows (Draft → Confirmed → Rejected)

### 2. Authentication and Authorization Scenarios

- **TestAuthenticationFlows**: Comprehensive authentication testing
  - Tests without authorization headers (should return 401)
  - Tests with invalid authorization tokens (should return 401)
  - Tests with valid authorization tokens (should pass authentication)
  - Tests different HTTP methods with authentication
  - Tests all endpoints for proper authentication requirements

### 3. Error Response Formats and Status Codes

- **TestErrorResponseFormats**: Tests error handling scenarios
  - Invalid UUID format validation (should return 400)
  - Non-existent MAWB Info UUID (should return 404)
  - Malformed JSON payloads (should return 400)
  - Missing required fields (should return 400)
  - Proper error message formatting

### 4. PDF Generation with File Validation

- **TestPDFGenerationEndpoints**: Tests PDF generation functionality
  - PDF generation for cargo manifest documents
  - PDF generation for draft MAWB documents
  - Content-Type header validation (application/pdf)
  - PDF magic number validation (%PDF)
  - File size validation (minimum size checks)
  - Error handling for PDF generation failures

## Test Structure

### Test Suite Setup

```go
type APIIntegrationTestSuite struct {
    suite.Suite
    server            *server.Server
    db                *pg.DB
    ctx               context.Context
    testToken         string
    testUserUUID      string
    testMAWBInfoUUIDs []string
    svcFactory        *factory.ServiceFactory
}
```

### Database Setup

- Automatic test database schema creation
- Foreign key relationships with CASCADE DELETE
- Test data cleanup after each test
- Transaction isolation for test reliability

### Authentication Setup

- Mock JWT token generation for testing
- User UUID context for authenticated requests
- Authorization header management

## Running the Tests

### Prerequisites

1. PostgreSQL test database
2. Environment variables set:
   ```bash
   export TEST_INTEGRATION=1
   export TEST_POSTGRESQL_HOST=localhost
   export TEST_POSTGRESQL_USER=postgres
   export TEST_POSTGRESQL_PASSWORD=your_password
   export TEST_POSTGRESQL_NAME=test_db
   export TEST_POSTGRESQL_PORT=5432
   export TEST_POSTGRESQL_SSLMODE=false
   ```

### Running All API Integration Tests

```bash
TEST_INTEGRATION=1 go test -v -run "TestAPIIntegrationTestSuite" -timeout 60s
```

### Running Specific Test Methods

```bash
# Test cargo manifest endpoints
TEST_INTEGRATION=1 go test -v -run "TestAPIIntegrationTestSuite/TestCargoManifestAPIEndpoints"

# Test authentication flows
TEST_INTEGRATION=1 go test -v -run "TestAPIIntegrationTestSuite/TestAuthenticationFlows"

# Test PDF generation
TEST_INTEGRATION=1 go test -v -run "TestAPIIntegrationTestSuite/TestPDFGenerationEndpoints"

# Test error handling
TEST_INTEGRATION=1 go test -v -run "TestAPIIntegrationTestSuite/TestErrorResponseFormats"
```

## Test Data Management

### Automatic Cleanup

- Test data is automatically cleaned up before and after each test
- Uses tracking slices to manage created UUIDs
- Cleanup follows reverse dependency order to avoid foreign key violations

### Test Isolation

- Each test creates its own test data
- Tests are isolated from each other
- Database transactions ensure data consistency

## Requirements Coverage

These integration tests fulfill the following requirements from the specification:

### Requirement 8.2 (Integration Tests)

✅ **Full request/response cycle tests for all endpoints**

- Complete CRUD operations for cargo manifest and draft MAWB
- All HTTP methods (GET, POST) tested
- Status transition endpoints tested

✅ **Authentication flows and authorization scenarios**

- Comprehensive authentication testing
- Invalid token handling
- Authorization header validation
- Multiple endpoint authentication testing

✅ **PDF generation endpoints with actual file validation**

- PDF content-type validation
- PDF magic number validation
- File size validation
- Error handling for PDF failures

✅ **Error response formats and status codes validation**

- HTTP status code validation (400, 401, 404, 500)
- Error message format validation
- Invalid input handling
- Malformed request testing

### Requirement 8.4 (API Testing)

✅ **Complete API endpoint coverage**

- All cargo manifest endpoints tested
- All draft MAWB endpoints tested
- Status management endpoints tested
- PDF generation endpoints tested

✅ **Request/response validation**

- JSON payload validation
- Response structure validation
- Data integrity validation
- Nested data structure validation

### Requirements 4.1-4.4 (PDF Generation)

✅ **PDF generation functionality**

- Cargo manifest PDF generation
- Draft MAWB PDF generation
- PDF header validation
- Content disposition validation

## Test Execution Flow

1. **Setup Phase**

   - Database connection establishment
   - Schema migration execution
   - Service factory initialization
   - Test server creation
   - JWT token generation

2. **Test Execution**

   - Individual test data creation
   - HTTP request execution
   - Response validation
   - Status code verification
   - Data integrity checks

3. **Cleanup Phase**
   - Test data removal
   - Database cleanup
   - Resource deallocation

## Error Handling

The tests include comprehensive error scenario coverage:

- **Network Errors**: Connection failures, timeouts
- **Authentication Errors**: Invalid tokens, missing headers
- **Validation Errors**: Invalid UUIDs, missing fields
- **Business Logic Errors**: Invalid status transitions
- **Database Errors**: Foreign key violations, constraint failures
- **PDF Generation Errors**: Template failures, font loading issues

## Performance Considerations

- Tests use connection pooling for database efficiency
- Cleanup operations are optimized for speed
- Test data is minimal but comprehensive
- Timeouts are set appropriately for CI/CD environments

## Continuous Integration

These tests are designed to run in CI/CD pipelines:

```yaml
# Example GitHub Actions configuration
- name: Run API Integration Tests
  env:
    TEST_INTEGRATION: 1
    TEST_POSTGRESQL_HOST: localhost
    TEST_POSTGRESQL_USER: postgres
    TEST_POSTGRESQL_PASSWORD: postgres
    TEST_POSTGRESQL_NAME: test_db
  run: |
    go test -v -run "TestAPIIntegrationTestSuite" -timeout 60s
```

## Maintenance

### Adding New Tests

1. Follow the existing test suite patterns
2. Ensure proper test data cleanup
3. Add appropriate assertions
4. Update this documentation

### Updating Existing Tests

1. Maintain backward compatibility
2. Update assertions as needed
3. Ensure test isolation is maintained
4. Verify cleanup operations work correctly

## Troubleshooting

### Common Issues

1. **Database Connection Failures**: Check environment variables and database availability
2. **Authentication Failures**: Verify JWT token generation and server configuration
3. **Test Data Conflicts**: Ensure cleanup operations are working correctly
4. **PDF Generation Failures**: Check font files and template availability

### Debug Mode

Run tests with verbose output for debugging:

```bash
TEST_INTEGRATION=1 go test -v -run "TestAPIIntegrationTestSuite" -count=1
```

The `-count=1` flag disables test caching for debugging purposes.
