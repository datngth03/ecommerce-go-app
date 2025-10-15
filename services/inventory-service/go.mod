module github.com/datngth03/ecommerce-go-app/services/inventory-service

go 1.24.0

toolchain go1.24.3

require (
	github.com/datngth03/ecommerce-go-app/proto v0.0.0
	github.com/datngth03/ecommerce-go-app/shared v0.0.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/rabbitmq/amqp091-go v1.9.0
	google.golang.org/grpc v1.76.0
	gorm.io/driver/postgres v1.5.4
	gorm.io/gorm v1.25.5
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.4.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	google.golang.org/protobuf v1.36.9 // indirect
)

replace github.com/datngth03/ecommerce-go-app/proto => ../../proto

replace github.com/datngth03/ecommerce-go-app/shared => ../../shared
