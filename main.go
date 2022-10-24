package main

import (
	"APKManagementServer/internal/logger"
	"APKManagementServer/xds"
	"APKManagementServer/xds/callbacks"
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"

	apkmgt_service "github.com/wso2/product-microgateway/adapter/pkg/discovery/api/wso2/discovery/service/apkmgt"
	wso2_server "github.com/wso2/product-microgateway/adapter/pkg/discovery/protocol/server/v3"
	"google.golang.org/grpc"
)

const (
	maxRandomInt             int = 999999999
	grpcMaxConcurrentStreams     = 1000000
	port                         = 18000
)

func main() {
	logger.LoggerServer.Info("Hello, world.")
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	go runServer()
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

func runServer() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	apkMgtAPIDsSrv := wso2_server.NewServer(ctx, xds.GetAPICache(), &callbacks.Callbacks{})

	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	grpcServer := grpc.NewServer(grpcOptions...)
	apkmgt_service.RegisterAPKMgtDiscoveryServiceServer(grpcServer, apkMgtAPIDsSrv)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.LoggerServer.Errorf("Error while listening on port: %s", port, err.Error())
	}

	logger.LoggerServer.Info("XDS server is starting.")
	if err = grpcServer.Serve(listener); err != nil {
		logger.LoggerServer.Error("Error while starting gRPC server.")
	}
}
