package main

import "github.com/davecheney/profile"

func main() {
	defer profile.Start(profile.CPUProfile).Stop()

	startServer()
}
