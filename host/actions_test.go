package host

import (
	"context"
	"errors"
	"testing"
)

func TestTypedActionRejectsUnknownFields(t *testing.T) {
	type request struct {
		Name string `json:"name"`
	}
	type response struct {
		Greeting string `json:"greeting"`
	}
	const name = "test.typed.strict"
	if err := RegisterAction(name, func(_ context.Context, _ *Session, request request) (response, error) {
		return response{Greeting: "hello " + request.Name}, nil
	}); err != nil {
		t.Fatalf("register action: %v", err)
	}

	result, actionErr := DispatchAction(context.Background(), newSession("typed"), name, map[string]any{"name": "Ada"})
	if actionErr != nil {
		t.Fatalf("dispatch action: %v", actionErr)
	}
	if result.(response).Greeting != "hello Ada" {
		t.Fatalf("unexpected response: %#v", result)
	}

	_, actionErr = DispatchAction(context.Background(), newSession("strict"), name, map[string]any{
		"name":  "Ada",
		"admin": true,
	})
	if actionErr == nil || actionErr.Code != "invalid_request" {
		t.Fatalf("unknown field was accepted: %#v", actionErr)
	}
}

func TestTypedActionAuthorizationHidesInternalError(t *testing.T) {
	type request struct {
		Owner string `json:"owner"`
	}
	const name = "test.typed.authorized"
	if err := RegisterAction(name,
		func(_ context.Context, _ *Session, request request) (request, error) {
			return request, nil
		},
		WithActionAuthorizer(func(_ context.Context, _ *Session, request request) error {
			if request.Owner != "allowed" {
				return errors.New("database policy detail")
			}
			return nil
		}),
	); err != nil {
		t.Fatalf("register action: %v", err)
	}

	_, actionErr := DispatchAction(context.Background(), newSession("denied"), name, map[string]any{"owner": "denied"})
	if actionErr == nil || actionErr.Code != "forbidden" || actionErr.Message != "action forbidden" {
		t.Fatalf("unexpected authorization response: %#v", actionErr)
	}
}

func TestTypedFormReturnsFieldErrors(t *testing.T) {
	type values struct {
		Email string `json:"email"`
	}
	type result struct {
		ID int `json:"id"`
	}
	const name = "test.form.validation"
	if err := RegisterForm(name,
		func(values values) FieldErrors {
			if values.Email == "" {
				return FieldErrors{"email": "required"}
			}
			return nil
		},
		func(_ context.Context, _ *Session, _ values) (result, error) {
			return result{ID: 7}, nil
		},
	); err != nil {
		t.Fatalf("register form: %v", err)
	}

	raw, actionErr := DispatchAction(context.Background(), newSession("form"), name, map[string]any{})
	if actionErr != nil {
		t.Fatalf("dispatch form: %v", actionErr)
	}
	response := raw.(FormResponse[result])
	if response.Valid || response.Fields["email"] != "required" {
		t.Fatalf("unexpected form response: %#v", response)
	}
}
