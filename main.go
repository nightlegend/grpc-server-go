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
	"github.com/nightlegend/grpc-server-go/api/test"
	grpclb "github.com/nightlegend/grpc-server-go/dns"
	pb "github.com/nightlegend/grpc-server-go/proto"
	"golang.org/x/net/context"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	httpAddress string
	etcdAddress string
	serv        = flag.String("service", "grpc_service", "service name")
	reg         *string
	num         = flag.Int("n", 1, "running process number")
)

func init() {
	httpAddress = os.Getenv("HTTP_ADDR")
	etcdAddress = os.Getenv("ETCD_ADDR")

	if httpAddress == "" {
		httpAddress = ":8080"
	}

	if etcdAddress == "" {
		fmt.Println("set to default endpoint")
		reg = flag.String("reg", "http://localhost:2379", "register etcd address")
	} else {
		reg = flag.String("reg", etcdAddress, "register etcd address")
	}

}

// Run a gRPC endpoint
func endpoint(i int) {
	fmt.Println("start endpoint: ", i)
	lis, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		glog.Fatalf("Could not listen: %v", err)
	}
	// fmt.Println(lis.Addr().(*net.TCPAddr).Port)
	err = grpclb.Register(*serv, "localhost", lis.Addr().(*net.TCPAddr).Port, *reg, time.Second*10, 15)
	if err != nil {
		panic(err)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		s := <-ch
		log.Printf("'%v' receive signal '%v'", lis.Addr().(*net.TCPAddr).Port, s)
		grpclb.UnRegister(*serv, *reg)
		os.Exit(1)
	}()
	log.Printf("starting hello service at %d", lis.Addr().(*net.TCPAddr).Port)

	s := googlegrpc.NewServer()
	pb.RegisterRouteGuideServer(s, &test.Server{ID: i})
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
	fmt.Printf("http server listener on: %v\n", httpAddress)
	return http.ListenAndServe(httpAddress, mux)
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
