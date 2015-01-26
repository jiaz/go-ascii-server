package main

import (
	"fmt"
	"github.com/davecheney/profile"
	"testing"
)

func TestImageEngineDecoding(t *testing.T) {
	defer profile.Start(profile.CPUProfile).Stop()
	frameChan := loadingMov("demo.m4v")
	j := 0
	for img := range frameChan {
		convertImageToHtml(img, 120)
		fmt.Println("Processing frame:", j)
		j += 1
	}
	fmt.Println("Total frame:", j)
}
