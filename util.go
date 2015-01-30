package main

import "C"

import (
	"log"
	"runtime/debug"
)

func fatal(err error) {
	debug.PrintStack()
	log.Fatal(err)
}

func checkRet(ret C.int, err error) error {
	if int(ret) == 0 {
		return nil
	} else {
		return err
	}
}
