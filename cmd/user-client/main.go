// cmd/client/user-client/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	user_client "github.com/datngth03/ecommerce-go-app/pkg/client/user" // Import mã gRPC đã tạo từ pkg/client/user
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure" // Dùng cho môi trường dev, không có TLS
)

func main() {
	// Địa chỉ của User Service (cùng cổng mà User Service đang lắng nghe)
	userSvcAddr := "localhost:50051"

	// Thiết lập kết nối gRPC đến User Service
	// Với insecure.NewCredentials(), chúng ta không sử dụng TLS/SSL cho kết nối,
	// điều này OK cho môi trường phát triển cục bộ nhưng KHÔNG NÊN dùng trong production.
	conn, err := grpc.Dial(userSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Không thể kết nối đến User Service: %v", err)
	}
	defer conn.Close() // Đảm bảo đóng kết nối khi hàm main kết thúc

	// Tạo một gRPC client cho UserService
	client := user_client.NewUserServiceClient(conn)

	// --- Test Case 1: Đăng ký người dùng mới ---
	fmt.Println("\n--- Đăng ký người dùng mới ---")
	registerReq := &user_client.RegisterUserRequest{
		Email:    "testuser@example.com",
		Password: "password123",
		FullName: "Test User One",
	}
	registerResp, err := client.RegisterUser(context.Background(), registerReq)
	if err != nil {
		log.Printf("Lỗi khi đăng ký người dùng: %v", err)
	} else {
		fmt.Printf("Đăng ký thành công! User ID: %s, Message: %s\n", registerResp.GetUserId(), registerResp.GetMessage())
	}

	// --- Test Case 2: Đăng ký người dùng đã tồn tại (sẽ lỗi) ---
	fmt.Println("\n--- Đăng ký người dùng đã tồn tại ---")
	registerResp2, err := client.RegisterUser(context.Background(), registerReq) // Dùng lại email
	if err != nil {
		log.Printf("Lỗi khi đăng ký người dùng (đã tồn tại): %v", err)
	} else {
		fmt.Printf("Đăng ký thành công (lỗi): User ID: %s, Message: %s\n", registerResp2.GetUserId(), registerResp2.GetMessage())
	}

	// --- Test Case 3: Đăng nhập người dùng ---
	fmt.Println("\n--- Đăng nhập người dùng ---")
	loginReq := &user_client.LoginUserRequest{
		Email:    "testuser@example.com",
		Password: "password123",
	}
	loginResp, err := client.LoginUser(context.Background(), loginReq)
	if err != nil {
		log.Printf("Lỗi khi đăng nhập người dùng: %v", err)
	} else {
		fmt.Printf("Đăng nhập thành công! User ID: %s, Access Token: %s\n", loginResp.GetUserId(), loginResp.GetAccessToken())
	}

	// --- Test Case 4: Lấy hồ sơ người dùng ---
	fmt.Println("\n--- Lấy hồ sơ người dùng ---")
	if registerResp != nil && registerResp.GetUserId() != "" {
		profileReq := &user_client.GetUserProfileRequest{
			UserId: registerResp.GetUserId(),
		}
		profileResp, err := client.GetUserProfile(context.Background(), profileReq)
		if err != nil {
			log.Printf("Lỗi khi lấy hồ sơ người dùng: %v", err)
		} else {
			fmt.Printf("Hồ sơ người dùng: ID: %s, Email: %s, Full Name: %s\n", profileResp.GetUserId(), profileResp.GetEmail(), profileResp.GetFullName())
		}
	} else {
		fmt.Println("Không thể lấy hồ sơ vì không có User ID từ đăng ký.")
	}

	// --- Test Case 5: Cập nhật hồ sơ người dùng ---
	fmt.Println("\n--- Cập nhật hồ sơ người dùng ---")
	if registerResp != nil && registerResp.GetUserId() != "" {
		updateReq := &user_client.UpdateUserProfileRequest{
			UserId:      registerResp.GetUserId(),
			FullName:    "Updated Test User",
			PhoneNumber: "0987654321",
			Address:     "123 Updated Street, City",
		}
		updateResp, err := client.UpdateUserProfile(context.Background(), updateReq)
		if err != nil {
			log.Printf("Lỗi khi cập nhật hồ sơ người dùng: %v", err)
		} else {
			fmt.Printf("Cập nhật hồ sơ thành công! ID: %s, Full Name mới: %s\n", updateResp.GetUserId(), updateResp.GetFullName())
		}
	} else {
		fmt.Println("Không thể cập nhật hồ sơ vì không có User ID từ đăng ký.")
	}

	fmt.Println("\nHoàn thành các test case.")
	time.Sleep(1 * time.Second) // Đợi một chút để log kịp in ra
}
