package xerr

import (
	"fmt"
)

type CodeError struct {
	Code uint32 `json:"code"`
	Msg  string `json:"msg"`
}

func (e *CodeError) Error() string {
	return fmt.Sprintf("ErrCode:%d，ErrMsg:%s", e.Code, e.Msg)
}

func New(code uint32, msg string) *CodeError {
	return &CodeError{Code: code, Msg: msg}
}
