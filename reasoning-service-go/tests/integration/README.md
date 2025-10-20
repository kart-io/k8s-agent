# Integration Tests

This directory contains integration tests for the Reasoning Service refactoring with LLM Proxy and LangChain.

## Directory Structure

```
tests/integration/
├── llm_proxy/              # LLM Proxy Adapter integration tests
│   └── adapter_integration_test.go
├── testutil/               # Testing utilities and mocks
│   ├── mock_llm.go        # Mock LLM implementation
│   └── helpers.go         # Test helper functions
└── testdata/              # Test data directory (created at runtime)
    ├── chroma/            # Chroma vector store data
    ├── learning/          # Learning system data
    └── logs/              # Test logs
```

## Running Tests

### Run all integration tests

```bash
go test -v ./tests/integration/...
```

### Run tests in short mode (skips all integration tests)

```bash
go test -v -short ./tests/integration/...
```

### Run specific test suite

```bash
go test -v ./tests/integration/llm_proxy/...
```

### Run with race detector

```bash
go test -v -race ./tests/integration/...
```

## Test Suites

### 1. LLM Proxy Adapter Tests (`llm_proxy/`)

Tests for the LLM Proxy Adapter implementation:

- **TestProxyAdapterIntegration**: End-to-end test loading config and creating adapter
- **TestProxyAdapterWithMock**: Tests using mock LLM for predictable responses
- **TestConfigValidation**: Validates test configuration loading
- **TestPerformanceBaseline**: Performance baseline for mock LLM operations

## Test Configuration

Integration tests use a dedicated test configuration file: `configs/config-test.yaml`

Key features of test config:

- **Different port**: Uses port 8083 to avoid conflicts with development server
- **Mock providers**: Configured with test API keys for mock server
- **New architecture enabled**: All new feature flags are enabled for testing
- **Reduced resources**: Smaller worker pools and timeouts for faster tests
- **Test data directory**: Uses `./tests/integration/testdata/` for all test data

## Mock LLM

The `testutil/mock_llm.go` provides a mock implementation of `gollm.LLM` interface for testing without real API calls:

### Creating a Mock LLM

```go
mockLLM := testutil.NewMockLLM()

// Add predefined responses
mockLLM.WithResponse("root cause", testutil.CreateMockRootCauseResponse())

// Simulate errors
mockLLM.WithError("simulated API error")
```

### Using Mock LLM

```go
ctx := context.Background()
prompt := testutil.NewMockPrompt("Analyze the root cause")
response, err := mockLLM.Generate(ctx, prompt)

// Check call history
callCount := mockLLM.GetCallCount()
lastCall := mockLLM.GetLastCall()
```

### Predefined Responses

The mock LLM includes factory functions for common response types:

- `CreateMockRootCauseResponse()`: Root cause analysis response
- `CreateMockDescriptionResponse()`: Failure description response
- `CreateMockRecommendationResponse()`: Recommendations response

## Test Helpers

The `testutil/helpers.go` provides common testing utilities:

### Load Test Configuration

```go
cfg := testutil.LoadTestConfig(t)
```

### Setup/Cleanup Test Data

```go
testDataDir := testutil.SetupTestDataDir(t)
defer testutil.CleanupTestDataDir(t)
```

### Skip Tests

```go
// Skip in short mode
testutil.SkipIfShort(t, "reason for skipping")

// Require environment variable
apiKey := testutil.RequireEnvVar(t, "OPENAI_API_KEY")
```

## Writing New Integration Tests

### Best Practices

1. **Use `-short` flag**: Use `testutil.SkipIfShort()` to allow skipping slow tests
2. **Use test config**: Load configuration with `testutil.LoadTestConfig(t)`
3. **Clean up**: Always clean up test data with defer
4. **Use mocks**: Use mock LLM for unit-style integration tests
5. **Real API tests**: Mark tests requiring real API keys with skip conditions

### Example Test Structure

```go
func TestMyFeature(t *testing.T) {
	testutil.SkipIfShort(t, "My feature integration test")

	cfg := testutil.LoadTestConfig(t)
	testDataDir := testutil.SetupTestDataDir(t)
	defer testutil.CleanupTestDataDir(t)

	// Your test code here
}
```

## Integration with Real APIs

To run tests with real LLM providers:

1. Set environment variables for API keys:

```bash
export OPENAI_API_KEY="your-key"
export GEMINI_API_KEY="your-key"
export DEEPSEEK_API_KEY="your-key"
```

2. Update `configs/config-test.yaml` to use real providers instead of mock

3. Run tests:

```bash
go test -v ./tests/integration/llm_proxy/...
```

## Performance Testing

The performance baseline test (`TestPerformanceBaseline`) measures mock LLM performance:

- Target: < 1ms average latency per mock call
- Current baseline: ~481ns per call
- Throughput: ~2M requests/second

For real API performance testing, create separate benchmarks:

```go
func BenchmarkRealAPI(b *testing.B) {
	// Setup with real API

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Your benchmark code
	}
}
```

## CI/CD Integration

### In CI Pipeline

Run tests in short mode to skip slow integration tests:

```bash
go test -v -short ./...
```

### Nightly Integration Tests

Run full integration tests with real APIs (requires API keys in CI):

```bash
export OPENAI_API_KEY="${OPENAI_API_KEY_SECRET}"
go test -v -timeout 30m ./tests/integration/...
```

## Troubleshooting

### Tests Skip with "no valid API keys"

This is expected when running without real API keys. The tests will use mock LLM instead.

### Test data directory errors

Make sure the project root is accessible and writable:

```bash
chmod -R 755 tests/integration/testdata
```

### Import errors

Run `go mod tidy` to ensure all dependencies are properly downloaded:

```bash
go mod tidy
```

## Future Enhancements

Planned improvements for integration tests:

- [ ] Mock HTTP server for end-to-end API testing
- [ ] Integration with Chroma vector store
- [ ] Integration with LangChain chains and agents
- [ ] Memory system integration tests
- [ ] Orchestrator integration tests
- [ ] Performance regression testing
