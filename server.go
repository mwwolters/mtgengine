package main

import (
	"flag"
	"fmt"
	"log"
	"mtgengine/engine"
	"mtgengine/srv"
	pb "mtgengine/proto"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	port = flag.Int("port", 9999, "server port")
	sets = flag.String("sets", "engine/AllSets.json", "path to the json for the sets")
	backupSets = flag.String("bsets", "/go/src/mtgengine/engine/AllSets.json", "path to the json for the sets")
)

func main() {
	flag.Parse()
	s, err := engine.LoadSets(*sets)
	if err != nil {
		fmt.Print("Error loading set, attempting backup\n")
		s, err = engine.LoadSets(*backupSets)
		if err != nil {
			fmt.Print("Error loading set, running in lame duck mode\n")
		}
	}
	flatSets := engine.FlattenSet(s)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("listening on %v", *port)
	grpcServer := grpc.NewServer()
	pb.RegisterCollectionEngineServer(grpcServer, service.NewService(flatSets))
	reflection.Register(grpcServer)
	grpcServer.Serve(lis)
}
