package pipeline

import "errors"

type SourceRejectedError struct {
	Reason string
}

func (e *SourceRejectedError) Error() string {
	return e.Reason
}

func (e *SourceRejectedError) Is(e2 error) bool {
	var err *SourceRejectedError
	ok := errors.As(e2, &err)
	return ok && e.Reason == err.Reason
}

type SourceSkippedError struct {
	Reason string
}

func (e *SourceSkippedError) Error() string {
	return e.Reason
}

func (e *SourceSkippedError) Is(e2 error) bool {
	var err *SourceSkippedError
	ok := errors.As(e2, &err)
	return ok && e.Reason == err.Reason
}

type UnsupportedCodecError struct {
	Codec string
}

func (e *UnsupportedCodecError) Error() string {
	return "unsupported codec: " + e.Codec
}

func (e *UnsupportedCodecError) Is(e2 error) bool {
	var err *UnsupportedCodecError
	ok := errors.As(e2, &err)
	return ok && e.Codec == err.Codec
}
