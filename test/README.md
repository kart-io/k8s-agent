# Test Infrastructure

This directory contains test infrastructure for the k8s-agent project.

## Directory Structure

```
test/
├── fixtures/       # Test fixtures (mock data, sample configs)
├── integration/    # Integration tests
├── e2e/           # End-to-end tests
└── testdata/      # Test data files
```

## Test Fixtures

The `fixtures/` directory contains reusable test data and utilities:

- Mock Kubernetes resources
- Sample configuration files
- Test database schemas
- Mock gRPC responses

## Integration Tests

The `integration/` directory contains tests that verify multiple components working together:

- Service-to-service communication
- Database integration
- Message queue integration
- External API integration

Run integration tests:
```bash
make go.test.integration
```

## End-to-End Tests

The `e2e/` directory contains full system tests:

- Complete workflow execution
- Multi-service scenarios
- Real Kubernetes cluster tests

Run E2E tests:
```bash
make test-e2e
```

## Test Data

The `testdata/` directory contains:

- Sample YAML/JSON files
- Test certificates
- Mock event data
- Test proto files

## Writing Tests

### Unit Tests

Place unit tests alongside your code:
```
pkg/mypackage/
├── myfile.go
└── myfile_test.go
```

### Integration Tests

Create integration tests in `test/integration/`:
```go
// +build integration

package integration

import "testing"

func TestServiceIntegration(t *testing.T) {
    // Test code
}
```

### E2E Tests

Create E2E tests in `test/e2e/`:
```go
// +build e2e

package e2e

import "testing"

func TestCompleteWorkflow(t *testing.T) {
    // Test code
}
```

## Test Utilities

Common test utilities are provided in:
- `test/fixtures/helpers.go` - Helper functions
- `test/fixtures/mocks.go` - Mock implementations
- `test/fixtures/builders.go` - Test data builders

## Best Practices

1. **Use table-driven tests** for multiple test cases
2. **Mock external dependencies** using interfaces
3. **Clean up resources** in test teardown
4. **Use meaningful test names** that describe the scenario
5. **Keep tests independent** - no test should depend on another
6. **Use test fixtures** to avoid duplication
7. **Add build tags** for integration and E2E tests

## Running Tests

```bash
# Run all unit tests
make go.test

# Run tests with coverage
make go.test.coverage

# Run integration tests
make go.test.integration

# Run E2E tests
make test-e2e

# Run specific test
go test -v ./pkg/mypackage -run TestMyFunction
```
