package main

import (
	"bookstoregrpc/database"
	"bookstoregrpc/pb"
	"bookstoregrpc/service"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	db := database.InitDB()
	ps := service.NewPostgresStore(db)
	BookServer := service.NewBookServer(ps)

	grpcServer := grpc.NewServer()
	pb.RegisterBookServiceServer(grpcServer, BookServer)

	address := "0.0.0.0:8080"
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("Cannot start server", err)
	}

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal("Cannot start server", err)
	}
}
