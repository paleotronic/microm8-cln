package exception

import "errors"

func NewESyntaxError(msg string) error {
	//panic(msg) 

	return errors.New(msg)
}
