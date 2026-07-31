// Package host provides server-side component sessions and transport support.
package host

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
)

// ActionHandler handles a typed client action.
type ActionHandler[Request, Response any] func(context.Context, *Session, Request) (Response, error)

// ActionAuthorizer can reject an action after the request is decoded.
type ActionAuthorizer[Request any] func(context.Context, *Session, Request) error

type actionConfig[Request any] struct {
	authorize ActionAuthorizer[Request]
}

// ActionOption configures a typed action.
type ActionOption[Request any] func(*actionConfig[Request])

// WithActionAuthorizer adds action-specific authorization.
func WithActionAuthorizer[Request any](authorize ActionAuthorizer[Request]) ActionOption[Request] {
	return func(config *actionConfig[Request]) {
		config.authorize = authorize
	}
}

type registeredAction interface {
	dispatch(context.Context, *Session, map[string]any) (any, *ActionError)
}

type typedAction[Request, Response any] struct {
	handler   ActionHandler[Request, Response]
	authorize ActionAuthorizer[Request]
}

func (a typedAction[Request, Response]) dispatch(ctx context.Context, session *Session, payload map[string]any) (response any, actionErr *ActionError) {
	defer func() {
		if recover() != nil {
			response = nil
			actionErr = NewActionError("action_failed", "action failed")
		}
	}()
	var request Request
	if err := decodeActionPayload(payload, &request); err != nil {
		return nil, &ActionError{Code: "invalid_request", Message: err.Error()}
	}
	if a.authorize != nil {
		if err := a.authorize(ctx, session, request); err != nil {
			return nil, publicActionError(err, "forbidden", "action forbidden")
		}
	}
	if a.handler == nil {
		return nil, NewActionError("not_implemented", "action handler is not configured")
	}
	response, err := a.handler(ctx, session, request)
	if err != nil {
		return nil, publicActionError(err, "action_failed", "action failed")
	}
	return response, nil
}

var actionRegistry = struct {
	sync.RWMutex
	actions map[string]registeredAction
}{actions: make(map[string]registeredAction)}

// RegisterAction registers a strict, typed SSC action.
func RegisterAction[Request, Response any](name string, handler ActionHandler[Request, Response], opts ...ActionOption[Request]) error {
	if name == "" {
		return errors.New("host: empty action name")
	}
	if handler == nil {
		return errors.New("host: nil action handler")
	}
	var config actionConfig[Request]
	for _, opt := range opts {
		opt(&config)
	}
	actionRegistry.Lock()
	defer actionRegistry.Unlock()
	if _, exists := actionRegistry.actions[name]; exists {
		return fmt.Errorf("host: action %q already registered", name)
	}
	actionRegistry.actions[name] = typedAction[Request, Response]{
		handler:   handler,
		authorize: config.authorize,
	}
	return nil
}

// DispatchAction decodes and executes a registered action.
func DispatchAction(ctx context.Context, session *Session, name string, payload map[string]any) (any, *ActionError) {
	actionRegistry.RLock()
	action := actionRegistry.actions[name]
	actionRegistry.RUnlock()
	if action == nil {
		return nil, NewActionError("action_not_found", "action not found")
	}
	return action.dispatch(ctx, session, payload)
}

func decodeActionPayload(payload map[string]any, target any) error {
	if payload == nil {
		payload = map[string]any{}
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return errors.New("request payload is not valid JSON")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("invalid request: multiple JSON values")
	}
	return nil
}

func publicActionError(err error, fallbackCode, fallbackMessage string) *ActionError {
	var actionErr *ActionError
	if errors.As(err, &actionErr) {
		return actionErr
	}
	return NewActionError(fallbackCode, fallbackMessage)
}

// FieldErrors maps form field names to validation messages.
type FieldErrors map[string]string

// FormResponse is returned by typed form actions.
type FormResponse[Response any] struct {
	Data   Response    `json:"data,omitempty"`
	Fields FieldErrors `json:"fields,omitempty"`
	Valid  bool        `json:"valid"`
}

// RegisterForm registers a typed action with field validation.
func RegisterForm[Values, Response any](
	name string,
	validate func(Values) FieldErrors,
	submit ActionHandler[Values, Response],
	opts ...ActionOption[Values],
) error {
	if submit == nil {
		return errors.New("host: nil form handler")
	}
	return RegisterAction(name, func(ctx context.Context, session *Session, values Values) (FormResponse[Response], error) {
		if validate != nil {
			if fields := validate(values); len(fields) > 0 {
				return FormResponse[Response]{Fields: fields}, nil
			}
		}
		data, err := submit(ctx, session, values)
		if err != nil {
			return FormResponse[Response]{}, err
		}
		return FormResponse[Response]{Data: data, Valid: true}, nil
	}, opts...)
}
