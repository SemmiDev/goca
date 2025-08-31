# Daftar target yang bukan file aktual
.PHONY: run-api swagger swagger-format up up-build down up-db up-redis down-db down-redis migrateup migrateup1 migratedown migratedown1 new_migration test lint tidy

# Variabel konfigurasi database
DB_URL=postgresql://user:password@localhost:5432/goca?sslmode=disable

# Menjalankan aplikasi Go secara lokal
run-api:
	go run ./cmd/api/main.go

# Menjalankan semua layanan (app, db, redis) menggunakan Docker Compose
up:
	docker compose up -d

# Menjalankan semua layanan dengan membangun ulang image
up-build:
	docker compose up -d --build

# Menghentikan dan menghapus semua container Docker Compose
down:
	docker compose down

down-volume:
	docker compose down --volumes

# Menjalankan container database (PostgreSQL)
up-db:
	docker compose up -d db

# Menjalankan container Redis
up-redis:
	docker compose up -d redis

# Menghentikan dan menghapus container database
down-db:
	docker compose down db

# Menghentikan dan menghapus container Redis
down-redis:
	docker compose down redis

# Menerapkan semua migrasi database
migrateup:
	migrate -path migrations -database "$(DB_URL)" -verbose up

# Menerapkan satu migrasi database
migrateup1:
	migrate -path migrations -database "$(DB_URL)" -verbose up 1

# Rollback semua migrasi database
migratedown:
	migrate -path migrations -database "$(DB_URL)" -verbose down

# Rollback satu migrasi database
migratedown1:
	migrate -path migrations -database "$(DB_URL)" -verbose down 1

# Membuat file migrasi baru (memerlukan variabel 'name')
new_migration:
	@if [ -z "$(name)" ]; then echo "Error: 'name' is required. Usage: make new_migration name=migration_name"; exit 1; fi
	migrate create -ext sql -dir migrations -seq $(name)

# Meng-generate dokumentasi Swagger
swagger:
	swag init -g cmd/api/main.go --parseInternal -o api/swagger

# Memformat anotasi Swagger
swagger-format:
	swag fmt

# Menjalankan semua unit test
test:
	go test ./... -v

# Menjalankan linter golangci-lint
lint:
	golangci-lint run -v

# Mengelola dependensi Go
tidy:
	go mod tidy
