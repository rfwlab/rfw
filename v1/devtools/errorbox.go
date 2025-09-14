package devtools

// runtimeError represents a captured runtime error.
type runtimeError struct {
	Message string
	Stack   string
	Path    string
}

var (
	errList []runtimeError
	errIdx  int = -1
)

const (
	maxRuntimeErrors = 15
)

// addRuntimeError appends a new runtime error and sets it as current.
func addRuntimeError(e runtimeError) {
	if len(errList) >= maxRuntimeErrors {
		return
	}
	errList = append(errList, e)
	errIdx = len(errList) - 1
}

// currentRuntimeError returns the active error.
func currentRuntimeError() (runtimeError, bool) {
	if errIdx < 0 || errIdx >= len(errList) {
		return runtimeError{}, false
	}
	return errList[errIdx], true
}

// prevRuntimeError moves to the previous error if available.
func prevRuntimeError() (runtimeError, bool) {
	if errIdx > 0 {
		errIdx--
	}
	return currentRuntimeError()
}

// nextRuntimeError moves to the next error if available.
func nextRuntimeError() (runtimeError, bool) {
	if errIdx < len(errList)-1 {
		errIdx++
	}
	return currentRuntimeError()
}

// resetRuntimeErrors clears all tracked errors.
func resetRuntimeErrors() {
	errList = nil
	errIdx = -1
}

// runtimeErrorCount returns the number of stored errors.
func runtimeErrorCount() int { return len(errList) }

// runtimeErrorIndex returns the current error index.
func runtimeErrorIndex() int { return errIdx }
