package parser

import (
	"github.com/glycerine/zygomys/zygo"
)

// SignalOk return a signal of success
func SignalOk(env *zygo.Zlisp) (zygo.Sexp, error) {
	signalOk := map[string]string{
		"status": "ok",
	}
	return ToSexp(env, signalOk), nil
}

// SignalErr return a signal of error
func SignalErr(env *zygo.Zlisp, err error) (zygo.Sexp, error) {
	signalErr := map[string]string{
		"status": "err",
	}
	return ToSexp(env, signalErr), err
}

// SignalWrongArgs return of a function with wrong args
func SignalWrongArgs() (zygo.Sexp, error) {
	return zygo.SexpNull, zygo.WrongNargs
}
