package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/prorok210/TestYourServer/app"
)

func main() {
	app.CreateAppWindow()
}
