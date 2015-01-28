package main

import (
	"log"
	"runtime/debug"
)

func fatal(err error) {
	debug.PrintStack()
	log.Fatal(err)
}
