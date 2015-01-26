package main

import (
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/websocket"
)

func getImageFrames() <-chan ImageFrame {
	frameChan := loadingMov("demo.m4v")
	fmt.Println("FrameChan:", frameChan)
	return frameChan
}

var videoBuffer []string

var cnt = 0

func preLoadVideo() {
	fmt.Println("warming up server...")
	videoBuffer = make([]string, 10000)
	imageFrames := getImageFrames()
	for imageFrame := range imageFrames {
		videoBuffer[cnt] = convertImageToHtml(imageFrame, 180)
		fmt.Println("Loading frame:", cnt)
		cnt++
	}
	fmt.Println("done")
}

func startServer() {
	preLoadVideo()

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, "Hello world\n")
	})

	server := websocket.Server{
		Handler: func(conn *websocket.Conn) {
			fmt.Println("on handler", conn)
			fmt.Println("Start streaming")
			for i := 0; i < cnt; i++ {
				fmt.Println("sending", i)
				websocket.Message.Send(conn, videoBuffer[i])
			}
			fmt.Println("Finished streaming")
		}}

	http.HandleFunc("/play", server.ServeHTTP)

	fmt.Println("Server started")

	err := http.ListenAndServe("127.0.0.1:5555", nil)
	if err != nil {
		fmt.Println("Listen failed")
	}

}
