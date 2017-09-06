package ideal_server_test
import (
	"testing"
	"github.com/viktorasm/redbutton/ideal_server"
	"google.golang.org/grpc"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"net"
)


func TestSimpleHello(t *testing.T) {
	lis, err := net.Listen("tcp",":50051")
	require.NoError(t, err)
	s := grpc.NewServer()
	ideal_server.RegisterIdealServerServer(s,ideal_server.NewServer())
	go s.Serve(lis)



	conn, err := grpc.Dial("0.0.0.0:50051", grpc.WithInsecure())
	require.NoError(t,err)
	defer conn.Close()
	c := ideal_server.NewIdealServerClient(conn)

	r, err := c.SayHello(context.Background(),&ideal_server.HelloRequest{Name: "whatever!"})
	require.NoError(t, err)
	require.Equal(t, "Hello whatever!", r.Message)


}