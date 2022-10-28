package main

import (
	"APKManagementServer/internal/logger"
	"APKManagementServer/internal/xds"
	"os"
	"os/signal"
)

func main() {
	logger.LoggerServer.Info("Hello, world.")
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	go xds.InitAPKMgtServer()
	go xds.InitDBConnection()
	// todo(amaliMatharaarachchi) watch data updates and update snapshot accordingly.
	go xds.FeedData()

OUTER:
	for {
		select {
		case s := <-sig:
			switch s {
			case os.Interrupt:
				break OUTER
			}
		}
	}
}
