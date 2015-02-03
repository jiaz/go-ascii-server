package main

import (
	"fmt"
	"testing"
)

func TestImageEngineDecoding(t *testing.T) {
	movie, err := loadMovie("resources/demo.m4v")
	if err != nil {
		t.Fatal("Cannot load movie")
	}
	converter, err := NewAsciiConverter(movie, 60)
	if err != nil {
		t.Fatal("Cannot create converter")
	}
	defer converter.Free()

	j := 0
	for img := range movie.ImageStream {
		t.Log("Processing frame:", j, &img)
		fmt.Println("w:", movie.Width, "h:", movie.Height)
		//v := string(processSimple(movie.Width, movie.Height, img))
		v, _ := converter.ConvertToAnsi(img)
		fmt.Println(v)
		j += 1
	}
	t.Log("Total frame:", j)
	// if movie.FrameCount != j {
	// 	t.Fatalf("Expected frame count: %d, but get: %d", movie.FrameCount, j)
	// }
}
