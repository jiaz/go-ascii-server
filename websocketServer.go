package main

import (
	"log"
	"net/http"
	"path/filepath"
	"text/template"

	"golang.org/x/net/websocket"
)

var (
	videoBuffer []string
	indexTmpl   *template.Template
	frameCount  int
)

func warmUp() {
	log.Println("warming up server...")

	movie, err := loadMovie(filepath.Join(config.ResourcesPath, "demo.m4v"))
	if err != nil {
		fatal(err)
	}
	converter, err := NewAsciiConverter(movie, 120)
	if err != nil {
		fatal(err)
	}
	defer converter.Free()

	videoBuffer = make([]string, movie.FrameCount)
	frameCount = 0
	for image := range movie.Images {
		videoBuffer[frameCount], err = converter.ConvertToHtml(image)
		if err != nil {
			fatal(err)
		}

		if frameCount%100 == 0 {
			log.Println("Loading frame:", frameCount)
		}
		frameCount++
		if frameCount == 2000 {
			break
		}
	}
	log.Println("warming up done")
}

func root(w http.ResponseWriter, r *http.Request) {
	var data = struct{ Host string }{config.WebsocketHost}
	err := indexTmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func registerHandler() {
	ws := websocket.Server{
		Handler: func(conn *websocket.Conn) {
			log.Println("Start streaming")
			for _, frame := range videoBuffer[0:frameCount] {
				websocket.Message.Send(conn, frame)
			}
			log.Println("Finished streaming")
		}}

	http.HandleFunc("/", root)
	http.Handle("/play", ws)
}

func bootstrap() {
	indexTmpl = template.Must(template.ParseFiles(filepath.Join(config.PublicPath, "index.html")))
}

func serve() {
	log.Fatal(http.ListenAndServe("0.0.0.0:"+config.ListenPort, nil))
}

func startServer() {
	loadConfig()
	bootstrap()
	warmUp()
	registerHandler()
	serve()
}
