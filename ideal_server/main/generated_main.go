// should be generated from spec
// takes care of starting the service
// takes care of finding the right port

package main
import (
	"google.golang.org/grpc"
	"net"
	"log"
	server "github.com/viktorasm/redbutton/ideal_server"
)


const (
	port = ":50051"
)


func main(){
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	server.RegisterIdealServerServer(s, server.NewServer())
	s.Serve(lis)
}
