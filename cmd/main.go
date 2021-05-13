package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ezhdanovskiy/wallets/internal/application"
)

func main() {
	app, err := application.NewApplication()
	if err != nil {
		log.Fatal(err)
	}

	go shutdownMonitor(app)

	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func shutdownMonitor(app *application.Application) {
	stopping := make(chan os.Signal)
	signal.Notify(stopping, os.Interrupt, syscall.SIGTERM)
	<-stopping

	app.Stop()
}
