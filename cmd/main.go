package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/akshaybt001/cart_service/db"
	"github.com/akshaybt001/cart_service/initializer"
	"github.com/akshaybt001/cart_service/service"
	"github.com/akshaybt001/proto_files/pb"
	"github.com/joho/godotenv"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatalf(err.Error())
	}
	prodConn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	defer func() {
		prodConn.Close()
	}()
	addr := os.Getenv("DB_KEY")

	DB, err := db.InitDB(addr)
	if err != nil {
		log.Fatal(err.Error())
	}
	prodRes := pb.NewProductServiceClient(prodConn)
	service.ProductClient = prodRes

	services := initializer.Initializer(DB)

	server := grpc.NewServer()

	pb.RegisterCartServiceServer(server, services)

	lis, err := net.Listen("tcp", ":8083")

	if err != nil {
		log.Fatalf("failed to listen on 8083 : %v", err)
	}
	healthService := &service.HealthChecker{}

	grpc_health_v1.RegisterHealthServer(server, healthService)

	tracer, closer := initTracer()

	defer closer.Close()

	service.RetrieveTracer(tracer)

	if err = server.Serve(lis); err != nil {
		log.Fatalf("Failed to connect on port 8083 : %v", err)
	}
}

func initTracer() (tracer opentracing.Tracer, closer io.Closer) {
	jaegerEndpoint := "http://localhost:14268/api/traces"

	cfg := &config.Configuration{
		ServiceName: "product-service",
		Sampler: &config.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:          true,
			CollectorEndpoint: jaegerEndpoint,
		},
	}

	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("updated")
	return
}
