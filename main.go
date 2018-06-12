package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	grpclb "github.com/nightlegend/grpc-server-go/dns"
	pb "github.com/nightlegend/grpc-server-go/proto"
	"golang.org/x/net/context"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	tcpPort  = 5000
	tcpAddr  = ":5000"
	httpAddr = ":8080"
	etcdAddr = "http://127.0.0.1:2379"
)

var (
	serv         = flag.String("service", "grpc_service", "service name")
	echoEndpoint = flag.String("echo_endpoint", tcpAddr, "endpoint of your service")
	reg          = flag.String("reg", etcdAddr, "register etcd address")
	endpointPort = flag.Int("port", tcpPort, "listening port")
	num          = flag.Int("n", 1, "running process number")
)

type server struct {
	id int
}

func (s *server) GetName(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	if req.Id == "1" {
		return &pb.Response{Name: "Peter"}, nil
	}
	fmt.Printf("%v: Receive is %s\n", time.Now(), req.Id)
	return &pb.Response{Name: "David Guo"}, nil
}

func (s *server) Echo(ctx context.Context, req *pb.StringMessage) (*pb.StringMessage, error) {
	return &pb.StringMessage{Value: "get value from grpc server"}, nil
}

func (s *server) GetInfo(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	fmt.Printf("the request from: %d\n", s.id)
	return &pb.Response{Name: "David Guo"}, nil
}

// Run a gRPC endpoint
func endpoint(i int) {
	fmt.Println("start endpoint: ", i)
	lis, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		glog.Fatalf("Could not listen: %v", err)
	}
	// fmt.Println(lis.Addr().(*net.TCPAddr).Port)
	err = grpclb.Register(*serv, "127.0.0.1", lis.Addr().(*net.TCPAddr).Port, *reg, time.Second*10, 15)
	if err != nil {
		panic(err)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		s := <-ch
		log.Printf("'%v' receive signal '%v'", lis.Addr().(*net.TCPAddr).Port, s)
		grpclb.UnRegister()
		os.Exit(1)
	}()
	log.Printf("starting hello service at %d", lis.Addr().(*net.TCPAddr).Port)

	s := googlegrpc.NewServer()
	pb.RegisterRouteGuideServer(s, &server{id: i})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		glog.Fatalf("failed to serve: %v", err)
	}
}

// Run starts a HTTP server and blocks while running if successful.
// The server will be shutdown when "ctx" is canceled.
func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	resolver := grpclb.NewResolver(*serv)
	bl := googlegrpc.RoundRobin(resolver)
	opts := []googlegrpc.DialOption{googlegrpc.WithInsecure(), googlegrpc.WithBalancer(bl)}
	err := pb.RegisterRouteGuideHandlerFromEndpoint(ctx, mux, *reg, opts)

	if err != nil {
		glog.Fatal(err)
		return err
	}

	return http.ListenAndServe(httpAddr, mux)
}

func main() {
	flag.Parse()
	defer glog.Flush()

	for i := 0; i < *num; i++ {
		go endpoint(i)
	}

	// start http server
	if err := run(); err != nil {
		glog.Fatal(err)
	}
}
