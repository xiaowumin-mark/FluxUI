# Form Validation Example

`examples/form_validation` is a React-style `RunElement` example for complex input and validation behavior.

## Run

```sh
go run ./examples/form_validation
```

## What It Covers

It exercises `TextFieldElement`, password mode, controlled values, programmatic state updates, submit validation, error feedback, and success feedback.

## P7 Smoke

| Step | Operation | Expected result |
| --- | --- | --- |
| FV-01 | Enter an empty username, invalid email, short password, and mismatched confirmation, then submit. | Validation fails and each invalid field appears in the error list. |
| FV-02 | Use the preset buttons for valid username, email, password, and matching confirmation, then submit. | Validation succeeds and the green success message is shown. |
| FV-03 | Edit a field after a submit result. | The submitted state clears and controlled text mirrors the latest user input. |
| FV-04 | Use clear/reset buttons. | Values reset without stale validation messages or duplicate change effects. |
