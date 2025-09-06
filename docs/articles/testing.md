# Testing

rfw uses Go's standard testing tools.

## Running the test suite

Run all tests with:

```bash
go test -v ./...
```

This executes tests for every package in the repository. Continuous integration runs the same command on each push and pull request, so keeping the suite green locally helps ensure CI passes.
