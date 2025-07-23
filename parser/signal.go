package parser

import (
	"github.com/glycerine/zygomys/zygo"
)

// ReturnOk return a signal of success
func SignalOk(env *zygo.Zlisp) (zygo.Sexp, error) {
	signalOk := map[string]string{
		"status": "ok",
	}
	return ToSexp(env, signalOk), nil
}

// ReturnErr return a signal of error
func SignalErr(env *zygo.Zlisp, err error) (zygo.Sexp, error) {
	signalErr := map[string]string{
		"status": "err",
	}
	return ToSexp(env, signalErr), err
}
