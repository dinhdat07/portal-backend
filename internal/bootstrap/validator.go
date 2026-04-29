package bootstrap

import "buf.build/go/protovalidate"

func NewValidator() protovalidate.Validator {
	v, err := protovalidate.New()
	if err != nil {
		panic(err)
	}
	return v
}
