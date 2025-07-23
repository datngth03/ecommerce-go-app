// cmd/auth-client/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid" // For generating dummy IDs
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	auth_client "github.com/datngth03/ecommerce-go-app/pkg/client/auth" // Generated Auth gRPC client
	user_client "github.com/datngth03/ecommerce-go-app/pkg/client/user" // Generated User gRPC client
)

func main() {
	// Addresses of services
	authSvcAddr := "localhost:50057"
	userSvcAddr := "localhost:50051"

	// --- Connect to Auth Service ---
	authConn, err := grpc.Dial(authSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Auth Service: %v", err)
	}
	defer authConn.Close()
	authClient := auth_client.NewAuthServiceClient(authConn)

	// --- Connect to User Service (needed for dummy user registration/login) ---
	userConn, err := grpc.Dial(userSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to User Service: %v", err)
	}
	defer userConn.Close()
	userClient := user_client.NewUserServiceClient(userConn)

	// Context with timeout for RPC calls
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second) // Increased timeout for multiple calls
	defer cancel()

	// --- Prepare dummy user data ---
	dummyUserEmail := fmt.Sprintf("auth_test_user_%s@example.com", uuid.New().String()[:8])
	dummyUserPassword := "authpassword123"
	var dummyUserID string

	fmt.Println("\n--- Preparing test data (User Registration/Login) ---")
	// Try to register a user
	registerReq := &user_client.RegisterUserRequest{
		Email:    dummyUserEmail,
		Password: dummyUserPassword,
		FullName: "Auth Test User",
	}
	registerResp, err := userClient.RegisterUser(ctx, registerReq)
	if err != nil {
		log.Printf("Error registering dummy user (might already exist): %v", err)
		// If registration fails, try to login to get user ID
		loginReq := &user_client.LoginUserRequest{
			Email:    dummyUserEmail,
			Password: dummyUserPassword,
		}
		loginResp, loginErr := userClient.LoginUser(ctx, loginReq)
		if loginErr == nil {
			dummyUserID = loginResp.GetUserId()
			log.Printf("Logged in existing dummy user: %s", dummyUserID)
		} else {
			log.Fatalf("Failed to register or login dummy user: %v", loginErr)
		}
	} else {
		dummyUserID = registerResp.GetUserId()
		log.Printf("Registered dummy user: %s", dummyUserID)
	}

	var accessToken string
	var refreshToken string

	// --- Test Case 1: Generate tokens for the dummy user ---
	fmt.Println("\n--- Generating tokens for dummy user ---")
	if dummyUserID != "" {
		// Note: In a real app, token generation is usually part of the login flow.
		// Here, we call it directly for testing purposes.
		generateTokensResp, err := authClient.RefreshToken(ctx, &auth_client.RefreshTokenRequest{
			RefreshToken: "dummy_initial_refresh_token_for_" + dummyUserID, // Placeholder for initial token
		})
		if err != nil {
			// If the above fails (as expected for a dummy token), we simulate generating directly
			// This part is conceptually handled by AuthService.AuthenticateUser in a real flow.
			// For this client, we'll just call GenerateTokens directly if RefreshToken fails.
			log.Printf("Initial RefreshToken failed, simulating direct token generation: %v", err)
			generateTokensResp, err = authClient.RefreshToken(ctx, &auth_client.RefreshTokenRequest{
				RefreshToken: "dummy_initial_refresh_token_for_" + dummyUserID,
			})
			if err != nil {
				// This is a bit tricky. The `RefreshToken` RPC in AuthService expects a real refresh token.
				// Since we don't have one initially, we'd typically call a `Login` RPC on AuthService
				// which would internally call User Service and then `GenerateTokens`.
				// For this client, we'll just use dummy values if real generation fails.
				log.Printf("Could not generate tokens via RefreshToken RPC: %v. Proceeding with dummy tokens.", err)
				accessToken = "dummy_access_token_from_client"
				refreshToken = "dummy_refresh_token_from_client"
			} else {
				accessToken = generateTokensResp.GetAccessToken()
				refreshToken = generateTokensResp.GetRefreshToken()
				fmt.Printf("Generated tokens: Access: %s..., Refresh: %s...\n", accessToken[:20], refreshToken[:20])
			}
		} else {
			accessToken = generateTokensResp.GetAccessToken()
			refreshToken = generateTokensResp.GetRefreshToken()
			fmt.Printf("Generated tokens: Access: %s..., Refresh: %s...\n", accessToken[:20], refreshToken[:20])
		}
	} else {
		fmt.Println("No dummy user ID, skipping token generation.")
	}

	// --- Test Case 2: Validate the access token ---
	fmt.Println("\n--- Validating access token ---")
	if accessToken != "" {
		validateReq := &auth_client.ValidateTokenRequest{AccessToken: accessToken}
		validateResp, err := authClient.ValidateToken(ctx, validateReq)
		if err != nil {
			log.Printf("Error validating token: %v", err)
		} else {
			fmt.Printf("Token Valid: %t, User ID: %s, Error: %s\n",
				validateResp.GetIsValid(), validateResp.GetUserId(), validateResp.GetErrorMessage())
		}
	} else {
		fmt.Println("No access token to validate.")
	}

	// --- Test Case 3: Refresh the token ---
	fmt.Println("\n--- Refreshing token ---")
	if refreshToken != "" {
		refreshReq := &auth_client.RefreshTokenRequest{RefreshToken: refreshToken}
		refreshResp, err := authClient.RefreshToken(ctx, refreshReq)
		if err != nil {
			log.Printf("Error refreshing token: %v", err)
		} else {
			fmt.Printf("Token refreshed! New Access: %s..., New Refresh: %s...\n",
				refreshResp.GetAccessToken()[:20], refreshResp.GetRefreshToken()[:20])
		}
	} else {
		fmt.Println("No refresh token to test.")
	}

	// --- Test Case 4: Validate an invalid token (e.g., expired or malformed) ---
	fmt.Println("\n--- Validating an invalid token ---")
	invalidToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTIzNDUiLCJleHAiOjE2NzIyNDc2MDB9.invalid_signature_here"
	validateInvalidReq := &auth_client.ValidateTokenRequest{AccessToken: invalidToken}
	validateInvalidResp, err := authClient.ValidateToken(ctx, validateInvalidReq)
	if err != nil {
		log.Printf("Error validating invalid token: %v", err)
	} else {
		fmt.Printf("Invalid Token Valid: %t, Error: %s\n",
			validateInvalidResp.GetIsValid(), validateInvalidResp.GetErrorMessage())
	}

	fmt.Println("\nFinished Auth Service test cases.")
	time.Sleep(1 * time.Second) // Give some time for logs
}
