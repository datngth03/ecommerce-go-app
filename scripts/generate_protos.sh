#!/bin/bash

# scripts/generate_protos.sh

# Thiết lập biến môi trường để đảm bảo các plugin được tìm thấy
# Đây là một cách để đảm bảo GOPATH/bin có trong PATH khi chạy script
export PATH=$PATH:$(go env GOPATH)/bin

# Thư mục gốc của dự án
PROJECT_ROOT=$(dirname "$(dirname "$(readlink -f "$0")")")

# Thư mục chứa các file .proto
PROTO_DIR="$PROJECT_ROOT/api/protobufs"

# Thư mục đích cho mã Go được tạo ra
# Theo cấu trúc đã đề xuất, mã client SDK sẽ nằm trong pkg/client
GEN_GO_DIR="$PROJECT_ROOT/pkg/client"

echo "Bắt đầu tạo mã Go từ các file .proto..."

# Tạo thư mục đích nếu nó chưa tồn tại
# Thêm lệnh này để đảm bảo thư mục pkg/client được tạo ra
mkdir -p "$GEN_GO_DIR"

# Duyệt qua tất cả các file .proto trong thư mục PROTO_DIR
# và tạo mã Go tương ứng
find "$PROTO_DIR" -name "*.proto" | while read proto_file; do
    echo "Đang xử lý: $proto_file"
    protoc \
        --proto_path="$PROTO_DIR" \
        --go_out="$GEN_GO_DIR" \
        --go_opt=paths=source_relative \
        --go-grpc_out="$GEN_GO_DIR" \
        --go-grpc_opt=paths=source_relative \
        "$proto_file"
done

echo "Đã tạo mã Go từ các file .proto thành công!"

# Chạy go mod tidy để cập nhật các dependencies mới (nếu có)
echo "Chạy go mod tidy để cập nhật dependencies..."
cd "$PROJECT_ROOT"
go mod tidy

echo "Hoàn tất quá trình tạo mã và cập nhật dependencies."

