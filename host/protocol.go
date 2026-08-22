package host

import "fmt"

// Inbound is a client-to-host SSC protocol message.
type Inbound struct {
	Component   string         `json:"component,omitempty"`
	Action      string         `json:"action,omitempty"`
	Control     string         `json:"control,omitempty"`
	ID          string         `json:"id,omitempty"`
	Payload     map[string]any `json:"payload,omitempty"`
	Sequence    uint64         `json:"sequence,omitempty"`
	Ack         uint64         `json:"ack,omitempty"`
	ResumeToken string         `json:"resumeToken,omitempty"`
}

// Outbound is a host-to-client SSC protocol message.
type Outbound struct {
	Component   string       `json:"component,omitempty"`
	Action      string       `json:"action,omitempty"`
	Control     string       `json:"control,omitempty"`
	ID          string       `json:"id,omitempty"`
	Payload     any          `json:"payload,omitempty"`
	Error       *ActionError `json:"error,omitempty"`
	Session     string       `json:"session,omitempty"`
	Sequence    uint64       `json:"sequence,omitempty"`
	Ack         uint64       `json:"ack,omitempty"`
	ResumeToken string       `json:"resumeToken,omitempty"`
}

// ActionError is a public, machine-readable action failure.
type ActionError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func (e *ActionError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewActionError creates a public action error safe to return to the client.
func NewActionError(code, message string) *ActionError {
	return &ActionError{Code: code, Message: message}
}
