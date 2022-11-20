package errors

import (
	"errors"
	"fmt"
	"log"
	"runtime"
)

type Error struct {
	// Op is the operation being performed
	// this can be named by convention as packageName.FunctionName
	// eg: var op = "errors.E"
	// for methods on structs this can be packageName.structName.methodName
	Op       Op
	Kind     Kind
	Err      error
	Location ErrLocation
}

type ErrLocation struct {
	File string
	Line int
}

func (e Error) Unwrap() error {
	return e.Err
}

func CustomErrToStringFormat(err error) string {
	var e *Error
	if !errors.As(err, &e) {
		return fmt.Sprintf("Err: %s", err.Error())
	}
	return fmt.Sprintf("Operation: %s, Kind: %s, Error: %s, File: %s Line: %d",
		string(e.Op),
		e.Kind.String(),
		e.Err.Error(),
		e.Location.File,
		e.Location.Line,
	)
}

type Unwrapper interface {
	Unwrap() error
}

var (
	_ error     = (*Error)(nil)
	_ Unwrapper = (*Error)(nil)
)

func (e *Error) Error() string {
	return e.Err.Error()
}

var _ error = (*Error)(nil)

type Op string

type Kind uint16

const (
	KindOther Kind = iota + 1
	KindInternal
	KindAPI
	KindBadInput
	KindNotFound
	KindInvalid
	KindNotAuthorized
	KindNotAllowed
	KindConflict
	KindExternalRPC
	KindNetwork
)

func (k Kind) String() string {
	switch k {
	case KindOther:
		return "other error"
	case KindInternal:
		return "internal error"
	case KindAPI:
		return "server API error"
	case KindBadInput:
		return "bad input"
	case KindNotFound:
		return "not found"
	case KindInvalid:
		return "invalid"
	case KindNotAuthorized:
		return "not authorized"
	case KindNotAllowed:
		return "not allowed"
	case KindConflict:
		return "data conflict"
	case KindExternalRPC:
		return "external rpc failed"
	case KindNetwork:
		return "network error"
	}
	return "unknown error kind"
}

func E(op Op, args ...interface{}) error {
	_, file, line, _ := runtime.Caller(1)
	e := &Error{
		Op: op,
		Location: ErrLocation{
			File: file,
			Line: line,
		},
	}
	for _, arg := range args {
		switch arg := arg.(type) {
		case Kind:
			e.Kind = arg
		case error:
			e.Err = arg
		case string:
			e.Err = errors.New(arg)
		default:
			log.Printf("errors.E: bad call from %s:%d:%v: %v", file, line, op, args)
		}
	}
	return e
}

func IsKind(want Kind, err error) bool {
	got := GetKind(err)
	return got == want
}

func Ops(err *Error) []Op {
	ops := []Op{err.Op}
	for {
		embeddedErr, ok := err.Err.(*Error)
		if !ok {
			break
		}

		ops = append(ops, embeddedErr.Op)
		err = embeddedErr
	}

	return ops
}

func GetKind(err error) Kind {
	var e *Error
	if !errors.As(err, &e) {
		return KindOther
	}
	if e.Kind != 0 {
		return e.Kind
	}
	return GetKind(e.Err)
}
