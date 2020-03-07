package utils

import (
	"os"
)

func ErrorSequence(fns ...func() error) error {
	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}

func FirstSuccess(fns ...func() error) (err error) {
	for _, fn := range fns {
		err = fn()
		if err == nil {
			break
		}
	}
	return err
}

func Shell() string {
	v := os.Getenv("SHELL")
	if v == "" {
		v = "sh"
	}
	return v
}
