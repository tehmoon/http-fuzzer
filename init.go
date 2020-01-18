package main

import (
	"os"
	"log"
	"syscall"
	"github.com/tehmoon/errors"
)

func init() {
	log.SetOutput(os.Stderr)

	var rLimit syscall.Rlimit

	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		err = errors.Wrap(err, "Error getting rLimits")
		panic(err)
	}

	rLimit.Cur = rLimit.Max

	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		err = errors.Wrap(err, "Error setting rLimits")
		panic(err)
	}
}
