package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/prorok210/TestYourServer/app"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	app.CreateAppWindow()
}
