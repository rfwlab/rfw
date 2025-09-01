# forms

Utilities for validating form fields using reactive signals. Validators return
`(bool, string)` where the string is an uppercase error code to handle in your
app.

| Function | Description |
| --- | --- |
| `Validate(field, validators...)` | Returns `(validSig, codeSig)` that update when `field` changes. |
| `Required(value)` | Validates non-empty strings (`IS_REQUIRED`). |
| `Numeric(value)` | Validates that a string contains only digits (`NOT_NUMERIC`). |

## Example

The example below combines `Required` and `Numeric` validators; the error signal
emits `IS_REQUIRED` when empty or `NOT_NUMERIC` for non-digit input.

@include:ExampleFrame:{code:"/examples/components/form_validation_component.go", uri:"/examples/form-validation"}
