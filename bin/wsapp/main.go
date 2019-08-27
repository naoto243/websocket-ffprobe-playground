package main

import (
	"github.com/naoto243/ffprobe-via-websocket/app"
	"os"
	"os/signal"
)

func main() {

	wsApp := app.NewWsApp()

	go func() {
		wsApp.Run()
	}()


	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	wsApp.Close()

}
