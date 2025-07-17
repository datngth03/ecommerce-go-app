#!/bin/bash

# scripts/generate_protos.sh
# Phiên bản này đọc tùy chọn go_package từ mỗi file .proto để xác định đúng thư mục đầu ra.

# Thiết lập biến môi trường để đảm bảo các plugin được tìm thấy
export PATH=$PATH:$(go env GOPATH)/bin

# Thư mục gốc của dự án (ví dụ: /d/WorkSpace/Personal/Go/ecommerce-go-app)
PROJECT_ROOT=$(dirname "$(dirname "$(readlink -f "$0")")")

# Module path của dự án (ví dụ: github.com/datngth03/ecommerce-go-app)
# Đọc từ go.mod để đảm bảo chính xác
PROJECT_MODULE_PATH=$(grep '^module' "$PROJECT_ROOT/go.mod" | awk '{print $2}')

# Thư mục chứa các file .proto (ví dụ: /d/.../ecommerce-go-app/api/protobufs)
PROTO_DIR="$PROJECT_ROOT/api/protobufs"

# Thư mục đích CƠ SỞ cho mã Go được tạo ra (ví dụ: /d/.../ecommerce-go-app/pkg/client)
GEN_GO_BASE_DIR="$PROJECT_ROOT/pkg/client"

echo "Bắt đầu tạo mã Go từ các file .proto..."
echo "Module gốc của dự án: $PROJECT_MODULE_PATH"
echo "Thư mục đầu ra Go cơ sở: $GEN_GO_BASE_DIR"

# Duyệt qua tất cả các file .proto trong thư mục PROTO_DIR
find "$PROTO_DIR" -name "*.proto" | while read proto_file; do
    echo "Đang xử lý: $proto_file"

    # 1. Đọc tùy chọn go_package từ file .proto
    # Ví dụ: option go_package = "github.com/datngth03/ecommerce-go-app/pkg/client/user;user_client";
    GO_PACKAGE_LINE=$(grep "option go_package" "$proto_file" | head -n 1)

    if [ -z "$GO_PACKAGE_LINE" ]; then
        echo "Cảnh báo: Không tìm thấy 'option go_package' trong $proto_file. Bỏ qua."
        continue
    fi

    # 2. Trích xuất đường dẫn import Go đầy đủ (ví dụ: github.com/datngth03/ecommerce-go-app/pkg/client/user)
    # Loại bỏ "option go_package = \"" và ";<package_name>\";"
    GO_IMPORT_PATH=$(echo "$GO_PACKAGE_LINE" | sed -E 's/option go_package = "([^;]+);.*";/\1/')

    if [ -z "$GO_IMPORT_PATH" ]; then
        echo "Lỗi: Không thể trích xuất đường dẫn import Go từ '$GO_PACKAGE_LINE' trong $proto_file. Bỏ qua."
        continue
    fi

    echo "  Đường dẫn import Go: $GO_IMPORT_PATH"

    # 3. Tính toán thư mục con tương đối từ GEN_GO_BASE_DIR
    # Ví dụ:
    # GO_IMPORT_PATH = github.com/datngth03/ecommerce-go-app/pkg/client/user
    # PROJECT_MODULE_PATH = github.com/datngth03/ecommerce-go-app
    # => remove PROJECT_MODULE_PATH/pkg/client/ from GO_IMPORT_PATH
    RELATIVE_SUB_DIR=$(echo "$GO_IMPORT_PATH" | sed -E "s|^${PROJECT_MODULE_PATH}/pkg/client/?||")

    # Nếu RELATIVE_SUB_DIR rỗng (ví dụ: cho common.proto nếu nó nằm trực tiếp trong pkg/client)
    # thì không thêm thư mục con.
    if [ "$RELATIVE_SUB_DIR" == "$GO_IMPORT_PATH" ]; then
        # Trường hợp đường dẫn import không chứa PROJECT_MODULE_PATH/pkg/client
        # Điều này có thể xảy ra nếu go_package không theo quy ước hoặc là common.proto
        # Xử lý đặc biệt cho trường hợp này, ví dụ: đặt trực tiếp vào GEN_GO_BASE_DIR
        GEN_GO_OUTPUT_DIR="$GEN_GO_BASE_DIR"
        echo "  Gói nằm trực tiếp trong pkg/client."
    else
        GEN_GO_OUTPUT_DIR="$GEN_GO_BASE_DIR/$RELATIVE_SUB_DIR"
        echo "  Thư mục con tương đối: $RELATIVE_SUB_DIR"
    fi

    # 4. Tạo thư mục đích cụ thể nếu nó chưa tồn tại
    mkdir -p "$GEN_GO_OUTPUT_DIR"
    echo "  Đang tạo thư mục đầu ra: $GEN_GO_OUTPUT_DIR"

    # 5. Chạy protoc để tạo mã Go và gRPC
    protoc \
        --proto_path="$PROTO_DIR" \
        --go_out="$GEN_GO_OUTPUT_DIR" \
        --go_opt=paths=source_relative \
        --go-grpc_out="$GEN_GO_OUTPUT_DIR" \
        --go-grpc_opt=paths=source_relative \
        "$proto_file"
done

echo "Đã tạo mã Go từ các file .proto thành công!"

# Chạy go mod tidy để cập nhật các dependencies mới (nếu có)
echo "Chạy go mod tidy để cập nhật dependencies..."
cd "$PROJECT_ROOT"
go mod tidy

echo "Hoàn tất quá trình tạo mã và cập nhật dependencies."

