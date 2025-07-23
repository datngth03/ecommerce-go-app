// cmd/notification-client/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid" // For generating dummy IDs
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	notification_client "github.com/datngth03/ecommerce-go-app/pkg/client/notification" // Generated Notification gRPC client
)

func main() {
	// Address of Notification Service
	notificationSvcAddr := "localhost:50058"

	// --- Connect to Notification Service ---
	notificationConn, err := grpc.Dial(notificationSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Notification Service: %v", err)
	}
	defer notificationConn.Close()
	notificationClient := notification_client.NewNotificationServiceClient(notificationConn)

	// Context with timeout for RPC calls
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// --- Prepare dummy data ---
	dummyUserID := uuid.New().String()
	dummyEmail := fmt.Sprintf("user_%s@example.com", uuid.New().String()[:8])
	dummyPhoneNumber := "+849" + uuid.New().String()[:8] // Dummy Vietnamese phone number prefix

	fmt.Println("\n--- Testing Notification Service ---")

	// --- Test Case 1: Send Email ---
	fmt.Println("\n--- Sending Email Notification ---")
	sendEmailReq := &notification_client.SendEmailRequest{
		RecipientEmail: dummyEmail,
		Subject:        "Welcome to Our E-commerce!",
		BodyHtml:       "<h1>Hello!</h1><p>Thank you for registering with us.</p>",
		BodyPlain:      "Hello! Thank you for registering with us.",
		UserId:         dummyUserID,
		TemplateName:   "welcome_email",
		TemplateData:   map[string]string{"username": "Test User"},
	}
	sendEmailResp, err := notificationClient.SendEmail(ctx, sendEmailReq)
	if err != nil {
		log.Printf("Error sending email: %v", err)
	} else {
		fmt.Printf("Email sent: Success: %t, Message: %s, Notification ID: %s\n",
			sendEmailResp.GetSuccess(), sendEmailResp.GetMessage(), sendEmailResp.GetNotificationId())
	}

	// --- Test Case 2: Send SMS ---
	fmt.Println("\n--- Sending SMS Notification ---")
	sendSMSReq := &notification_client.SendSMSRequest{
		RecipientPhoneNumber: dummyPhoneNumber,
		Message:              "Your order #12345 has been shipped!",
		UserId:               dummyUserID,
	}
	sendSMSResp, err := notificationClient.SendSMS(ctx, sendSMSReq)
	if err != nil {
		log.Printf("Error sending SMS: %v", err)
	} else {
		fmt.Printf("SMS sent: Success: %t, Message: %s, Notification ID: %s\n",
			sendSMSResp.GetSuccess(), sendSMSResp.GetMessage(), sendSMSResp.GetNotificationId())
	}

	// --- Test Case 3: Send Push Notification ---
	fmt.Println("\n--- Sending Push Notification ---")
	sendPushReq := &notification_client.SendPushNotificationRequest{
		UserId:      dummyUserID,
		DeviceToken: "dummy_device_token_12345", // In real app, this would be a real device token
		Title:       "New Product Alert!",
		Body:        "Check out our latest arrivals!",
		Data:        map[string]string{"product_id": uuid.New().String()},
	}
	sendPushResp, err := notificationClient.SendPushNotification(ctx, sendPushReq)
	if err != nil {
		log.Printf("Error sending push notification: %v", err)
	} else {
		fmt.Printf("Push Notification sent: Success: %t, Message: %s, Notification ID: %s\n",
			sendPushResp.GetSuccess(), sendPushResp.GetMessage(), sendPushResp.GetNotificationId())
	}

	fmt.Println("\nFinished Notification Service test cases.")
	time.Sleep(1 * time.Second) // Give some time for logs
}
