// cmd/search-service/main.go
package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv" // Đã sửa đường dẫn import
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/elastic/go-elasticsearch/v8" // Elasticsearch client

	"github.com/datngth03/ecommerce-go-app/internal/search/application"
	search_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/search/delivery/grpc" // Alias để tránh xung đột tên
	"github.com/datngth03/ecommerce-go-app/internal/search/infrastructure/repository"
	search_client "github.com/datngth03/ecommerce-go-app/pkg/client/search" // Generated gRPC client
)

// main là hàm entry point của Search Service.
func main() {
	// Tải biến môi trường từ file .env
	if err := godotenv.Load(); err != nil {
		log.Printf("Không tìm thấy file .env, đang đọc từ biến môi trường hệ thống: %v", err)
	}

	esAddress := os.Getenv("ELASTICSEARCH_ADDR")
	if esAddress == "" {
		esAddress = "http://localhost:9200" // Địa chỉ mặc định của Elasticsearch
	}

	grpcPort := os.Getenv("GRPC_PORT_SEARCH")
	if grpcPort == "" {
		grpcPort = "50061" // Cổng mặc định cho Search Service
	}

	// Khởi tạo Elasticsearch client
	esConfig := elasticsearch.Config{
		Addresses: []string{esAddress},
		// Tắt xác thực SSL nếu đang dùng HTTP và không có chứng chỉ (chỉ dev)
		// Transport: &http.Transport{
		// 	TLSClientConfig: &tls.Config{
		// 		InsecureSkipVerify: true,
		// 	},
		// },
	}
	esClient, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		log.Fatalf("Không thể khởi tạo Elasticsearch client: %v", err)
	}

	// Kiểm tra kết nối Elasticsearch
	res, err := esClient.Info()
	if err != nil {
		log.Fatalf("Lỗi khi lấy thông tin Elasticsearch: %v", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		log.Fatalf("Lỗi kết nối Elasticsearch: %s", res.String())
	}
	log.Printf("Đã kết nối thành công đến Elasticsearch: %s", res.String())

	// Khởi tạo Search Repository
	searchRepo := repository.NewElasticsearchProductRepository(esClient)

	// Khởi tạo Application Service
	searchService := application.NewSearchService(searchRepo)

	// Khởi tạo gRPC Server
	grpcServer := search_grpc_delivery.NewSearchGRPCServer(searchService)

	// Tạo một listener trên cổng đã định nghĩa
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Không thể lắng nghe cổng %s: %v", grpcPort, err)
	}

	// Tạo một instance của gRPC server
	s := grpc.NewServer()

	// Đăng ký SearchGRPCServer với gRPC server
	search_client.RegisterSearchServiceServer(s, grpcServer)

	// Đăng ký reflection service.
	reflection.Register(s)

	log.Printf("Search Service (gRPC) đang lắng nghe tại cổng :%s...", grpcPort)

	// Bắt đầu gRPC server
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Không thể phục vụ gRPC server: %v", err)
	}
}
