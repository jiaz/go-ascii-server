package main

import "testing"

func TestImageEngineDecoding(t *testing.T) {
	movie, err := loadMovie("resources/demo.m4v")
	if err != nil {
		t.Fatal("Cannot load movie")
	}
	converter, err := NewAsciiConverter(movie, 80)
	if err != nil {
		t.Fatal("Cannot create converter")
	}
	defer converter.Free()

	j := 0
	for img := range movie.ImageStream {
		t.Log("Processing frame:", j, &img)
		j += 1
	}
	t.Log("Total frame:", j)
	// if movie.FrameCount != j {
	// 	t.Fatalf("Expected frame count: %d, but get: %d", movie.FrameCount, j)
	// }
}
