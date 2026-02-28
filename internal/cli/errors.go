package cli

import (
	"errors"
	"fmt"
)

const (
	ExitOK         = 0
	ExitUserError  = 2
	ExitDependency = 3
	ExitRuntime    = 4
)

type AppError struct {
	Code int
	Msg  string
	Err  error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return e.Msg
	}
	if e.Msg == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("%s: %v", e.Msg, e.Err)
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func userError(msg string, err error) error {
	return &AppError{Code: ExitUserError, Msg: msg, Err: err}
}

func dependencyError(msg string, err error) error {
	return &AppError{Code: ExitDependency, Msg: msg, Err: err}
}

func runtimeError(msg string, err error) error {
	return &AppError{Code: ExitRuntime, Msg: msg, Err: err}
}

func errorExitCode(err error) int {
	if err == nil {
		return ExitOK
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		if appErr.Code != 0 {
			return appErr.Code
		}
	}
	return ExitRuntime
}
