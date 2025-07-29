package logger

import (
	"log" // Vẫn giữ để log lỗi khởi tạo logger (nếu logger chính chưa sẵn sàng)
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger là một instance của zap.Logger.
// Đây là logger toàn cục mà các service sẽ sử dụng.
var Logger *zap.Logger

// InitLogger khởi tạo logger toàn cục cho ứng dụng.
// Nó đọc cấp độ log từ biến môi trường và thiết lập cấu hình.
func InitLogger() {
	var config zap.Config
	logLevel := os.Getenv("LOG_LEVEL")

	switch logLevel {
	case "DEBUG":
		config = zap.NewDevelopmentConfig()
		// Development config thường có output dễ đọc hơn trên console.
		// Có thể thêm màu sắc nếu terminal hỗ trợ.
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	case "INFO":
		config = zap.NewProductionConfig() // Production config mặc định là INFO
	case "WARN":
		config = zap.NewProductionConfig()
		config.Level.SetLevel(zapcore.WarnLevel)
	case "ERROR":
		config = zap.NewProductionConfig()
		config.Level.SetLevel(zapcore.ErrorLevel)
	case "FATAL":
		config = zap.NewProductionConfig()
		config.Level.SetLevel(zapcore.FatalLevel)
	default:
		// Mặc định là INFO trong production config nếu LOG_LEVEL không được đặt
		config = zap.NewProductionConfig()
	}

	// Đặt lại encoder config để đảm bảo định dạng JSON (production) đẹp và dễ đọc
	// hoặc định dạng console rõ ràng (development).
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // Định dạng thời gian chuẩn ISO
	config.EncoderConfig.TimeKey = "timestamp"                   // Tên trường cho timestamp
	// config.EncoderConfig.EncodeLevel đã được đặt ở switch case
	config.EncoderConfig.LevelKey = "level"           // Tên trường cho cấp độ log
	config.EncoderConfig.MessageKey = "message"       // Tên trường cho tin nhắn log
	config.EncoderConfig.CallerKey = "caller"         // Bao gồm thông tin file/line gọi log
	config.EncoderConfig.StacktraceKey = "stacktrace" // Bao gồm stacktrace cho lỗi

	var err error
	Logger, err = config.Build()
	if err != nil {
		// Nếu logger không thể khởi tạo, log ra lỗi Fatal bằng logger tiêu chuẩn và thoát
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	Logger.Info("Logger initialized successfully.") // Dòng log đầu tiên bằng logger mới
}

// SyncLogger đảm bảo tất cả các log đang chờ được ghi trước khi thoát ứng dụng.
// Điều này rất quan trọng để tránh mất log khi ứng dụng tắt đột ngột.
func SyncLogger() {
	if Logger != nil {
		// Sync có thể trả về lỗi, nhưng thường có thể bỏ qua khi graceful shutdown
		err := Logger.Sync()
		if err != nil && err.Error() != "sync /dev/stderr: invalid argument" && err.Error() != "sync /dev/stdout: invalid argument" {
			// Bỏ qua các lỗi sync phổ biến trên một số môi trường Docker/Terminal khi stdout/stderr không phải là terminal thực
			log.Printf("Warning: Failed to sync logger: %v", err)
		}
	}
}
