package genfn

import "jsouthworth.net/go/dyn"

func MakeGeneric(fn interface{}) func(...interface{}) interface{} {
	switch v := fn.(type) {
	case func(...interface{}) interface{}:
		return v
	default:
		return func(args ...interface{}) interface{} {
			return dyn.Apply(v, args...)
		}
	}
}
