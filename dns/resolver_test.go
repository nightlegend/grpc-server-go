package dns

import (
	"flag"
	"testing"
)

var (
	serv = flag.String("service", "grpc_service", "service name")
)

func TestResolve(t *testing.T) {
	resolver := NewResolver(*serv)
	if resolver.serviceName != "grpc_service" {
		t.Fatalf("Expected service name is `grpc_service`, got %s", resolver.serviceName)
	}
}
