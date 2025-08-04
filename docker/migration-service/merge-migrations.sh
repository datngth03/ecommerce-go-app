#!/bin/sh

set -e

SOURCE_DIR="./migrations"
TARGET_DIR="./dist/migrations"

echo "--- Copying all migration files into $TARGET_DIR ---"

# Xóa thư mục đích nếu đã tồn tại
if [ -d "$TARGET_DIR" ]; then
    echo "  -> Clearing old directory: $TARGET_DIR"
    rm -rf "$TARGET_DIR"
fi
mkdir -p "$TARGET_DIR"

# Copy tất cả các file .up.sql và .down.sql từ mọi subfolder
find "$SOURCE_DIR" -type f \( -name "*.up.sql" -o -name "*.down.sql" \) -exec cp {} "$TARGET_DIR/" \;

echo "--- Done. All migrations copied to $TARGET_DIR ---"
