# Go Clean Architecture Template

Sebuah *boilerplate* atau templat untuk layanan RESTful API berbasis Go yang mengikuti prinsip **Clean Architecture**. Desain ini bertujuan untuk menciptakan struktur proyek yang modular, mudah diuji (*testable*), dapat diskalakan (*scalable*), dan mudah dipelihara (*maintainable*). Proyek ini sudah menyertakan fitur inti seperti manajemen **Pengguna** (*User*) dengan autentikasi JWT dan **Catatan** (*Notes*), serta dukungan untuk *observability*, pemrosesan latar belakang (*background processing*), dan *containerization*.

Proyek ini dirancang untuk para pengembang (*developer*) yang ingin memulai proyek API Go dengan fondasi yang kokoh, sambil menerapkan praktik terbaik (*best practices*) seperti *dependency injection*, *separation of concerns*, dan pengujian (*testing*).

-----

## ✨ Fitur Utama

  - 🏗️ **Clean Architecture**: Pemisahan lapisan yang jelas antara domain (entitas & logika bisnis), *use case* (layanan), *delivery* (*handler*), dan infrastruktur (*repository*, basis data, dll.) untuk memudahkan pemeliharaan dan pengujian.
  - 🔐 **Autentikasi & Otorisasi**: Registrasi, *login*, dan proteksi *endpoint* menggunakan JWT. Termasuk fitur lupa sandi (*forgot password*) dan verifikasi email.
  - 📝 **Manajemen Catatan**: Operasi CRUD lengkap untuk catatan (*notes*) yang terikat pada setiap pengguna.
  - ⚙️ **Pekerja Latar Belakang (*Background Worker*)**: Proses asinkron (*asynchronous*) menggunakan Redis sebagai perantara pesan (*message broker*) untuk tugas seperti pengiriman email (melalui SMTP).
  - 📈 **Pembatasan Laju & Keamanan**: Pembatasan permintaan (*rate limiting*) dengan Redis, *middleware* untuk CORS, ID permintaan (*request ID*), pemulihan dari *panic*, dan *header* keamanan (*security headers*).
  - 📜 **Logging Terstruktur**: Menggunakan Zap untuk *logging* terstruktur dalam format JSON yang mudah dianalisis, dengan integrasi *tracing*.
  - 🔍 **Observability**: Dukungan untuk pemantauan (*monitoring*) dengan Prometheus, Grafana, Loki, Tempo, dan OpenTelemetry (OTel) melalui konfigurasi di direktori `scripts/`.
  - 📦 **Containerized & Orchestrated**: Siap untuk di-*deploy* dengan Docker dan Docker Compose, termasuk pengaturan untuk basis data, *cache*, dan perangkat *observability*.
  - 📄 **Dokumentasi API**: Dokumentasi API (OpenAPI/Swagger) yang dibuat secara otomatis dari anotasi kode, lengkap dengan contoh permintaan dan respons.
  - ✅ **Pengujian & Linting**: *Unit test* untuk komponen-komponen kunci, serta *golangci-lint* untuk memastikan kualitas kode.
  - 📧 **Templat Email**: Templat email yang disematkan (*embedded*) untuk verifikasi dan lupa sandi.
  - 🔑 **Utilitas Pendukung**: Termasuk *cache* (Redis), enkripsi (AES), *hashing* kata sandi (*password hashing* dengan Bcrypt), pembuatan OTP, penjadwal (*scheduler* dengan GoCron), dan validator (Go-Playground).

-----

## 🛠️ Rangkaian Teknologi (*Tech Stack*)

| Kategori | Teknologi/Alat |
|---|---|
| **Bahasa** | Go (v1.25.0+) |
| **Framework HTTP** | Fiber (untuk *routing* dan *middleware*) |
| **Basis Data** | PostgreSQL (dengan pelacakan *query*) |
| **Cache & Broker** | Redis (untuk *caching*, *rate limiting*, *worker*) |
| **Autentikasi** | JWT, Bcrypt |
| **Email** | SMTP dengan templat |
| **Observability** | OpenTelemetry, Prometheus, Grafana, Loki, Tempo, Promtail |
| **Migrasi DB** | golang-migrate |
| **Dokumentasi** | Swaggo (Swagger/OpenAPI) |
| **Linting & Pengujian** | golangci-lint, Go testing framework |
| **Containerization** | Docker, Docker Compose |
| **Lainnya** | GoCron (*scheduler*), Redis (*worker distributor*) |

-----

## 🏛️ Struktur Proyek

Struktur proyek mengikuti standar konvensi Go dengan adaptasi dari Clean Architecture. Setiap modul (seperti `user` dan `note`) memiliki lapisannya sendiri untuk memastikan independensi.

Berikut adalah struktur direktori proyek:

```
.
├── api                  # Definisi API (proto, dokumen Swagger)
│   ├── proto
│   └── swagger          # Berkas Swagger yang dihasilkan
├── cmd                  # Titik masuk (entrypoint) aplikasi
│   └── api
│       └── main.go      # Berkas utama untuk menjalankan server
├── compose.yaml         # Konfigurasi Docker Compose untuk pengembangan/produksi
├── Dockerfile           # Instruksi untuk membangun image Docker
├── go.mod & go.sum      # Dependensi Go
├── internal             # Logika bisnis inti (tidak untuk diimpor dari luar)
│   ├── app.go           # Inisialisasi aplikasi utama
│   ├── config           # Manajemen konfigurasi (.env, konstanta)
│   ├── modules          # Modul domain per fitur
│   │   ├── note         # Modul notes: DTO, entitas, handler, repo, service
│   │   └── user         # Modul user: DTO, entitas, handler, repo, service
│   ├── pkg              # Paket utilitas yang digunakan bersama (apperror, cache, dll.)
│   │   ├── apperror     # Penanganan error kustom
│   │   ├── assets       # Aset yang disematkan (templat email)
│   │   ├── cache        # Cache Redis
│   │   ├── database     # Koneksi Postgres & pelacakan query
│   │   ├── email        # Pengirim email SMTP
│   │   ├── encoding     # Encoding Base64
│   │   ├── encryption   # Enkripsi AES
│   │   ├── logger       # Logger Zap
│   │   ├── observability# Middleware & utilitas OTel
│   │   ├── password     # Hashing Bcrypt
│   │   ├── random       # Pembuat OTP & string acak
│   │   ├── ratelimit    # Pembatas permintaan berbasis Redis
│   │   ├── request      # Penyaringan permintaan
│   │   ├── response     # Helper untuk respons API
│   │   ├── scheduler    # Penjadwal GoCron
│   │   ├── token        # Manajemen token JWT
│   │   ├── validator    # Validator Go-Playground
│   │   └── worker       # Pekerja latar belakang (distributor Redis, tugas)
│   └── server           # Pengaturan server (rute API, middleware, dll.)
├── logs                 # Berkas log (dihasilkan)
├── Makefile             # Otomatisasi tugas (build, test, migrate, dll.)
├── migrations           # Berkas migrasi SQL untuk skema DB
├── playground.http      # Berkas untuk pengujian API (misalnya, via VSCode REST Client)
├── README.md            # Dokumentasi ini
└── scripts              # Konfigurasi untuk alat observability (Grafana, dll.)
```

### Diagram Arsitektur Sederhana

Berikut adalah representasi sederhana dari arsitektur proyek ini:

```
+-------------------+     +-------------------+     +-------------------+
|      Delivery     |     |      Use Case     |     |       Domain      |
|   (Handler, API)  |<--->|     (Service)     |<--->|     (Entitas)     |
+-------------------+     +-------------------+     +-------------------+
         ^                         ^
         |                         |
         v                         v
+-------------------+     +-------------------+
|   Infrastruktur   |     |     Eksternal     |
|(Repositori, DB, ...)|   | (Email, Observ.)  |
+-------------------+     +-------------------+
```

-----

## 🚀 Memulai

### Prasyarat

  - Git
  - Go (v1.25.0+)
  - Docker & Docker Compose
  - Make

Instal beberapa alat Go tambahan yang diperlukan:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 1\. Salin (*Clone*) Repositori

```bash
git clone https://github.com/semmidev/goca.git
cd goca
```

### 2\. Konfigurasi Lingkungan (*Environment*)

Salin `.env.example` menjadi `.env`, lalu sesuaikan isinya:

```bash
cp .env.example .env
```

Isi variabel yang diperlukan seperti `DATABASE_URL`, `REDIS_ADDR`, `JWT_SECRET_KEY`, dan konfigurasi `SMTP_*` untuk email.

### 3\. Jalankan dengan Docker (Direkomendasikan)

```bash
make up
```

Perintah ini akan menjalankan aplikasi, Postgres, Redis, dan semua layanan *observability*. Aplikasi dapat diakses di `http://localhost:8080`.

Untuk menghentikan semua layanan:

```bash
make down
```

### 4\. Jalankan Secara Lokal (Tanpa Docker)

1.  Pastikan Postgres & Redis sudah berjalan di sistem Anda.
2.  `go mod tidy`
3.  `make migrateup`
4.  `make run-api`

Aplikasi akan berjalan pada *port* yang ditentukan di berkas `.env` (standarnya: 8080).

### Pengaturan Observability

  - **Grafana**: `http://localhost:3000` (user/pass standar: admin/admin)
  - **Prometheus**: `http://localhost:9090`
  - **Loki & Tempo** juga tersedia untuk logging dan tracing.

-----

## ⚙️ Perintah Makefile

| Perintah | Deskripsi |
|---|---|
| `make up` | Menjalankan semua *container* Docker |
| `make down` | Menghentikan & menghapus *container* |
| `make run-api` | Menjalankan aplikasi secara lokal |
| `make test` | Menjalankan *unit test* |
| `make swagger` | Membuat dokumen Swagger |
| `make migrateup` | Menerapkan migrasi basis data |
| `make migratedown` | Membatalkan migrasi terakhir |
| `make new_migration name=xyz` | Membuat berkas migrasi baru |
| `make lint` | Menjalankan *linter* |

-----

## 📄 Dokumentasi API

Setelah aplikasi berjalan, dokumentasi API dapat diakses melalui `http://localhost:8080/swagger/index.html`. Untuk memperbarui dokumen, jalankan perintah `make swagger`.

-----

## 🤝 Berkontribusi

*Fork* repositori ini, buat *branch* baru, lakukan perubahan dan *commit*, lalu buat *Pull Request*. Gunakan *Issues* untuk memulai diskusi.

-----

## 📜 Lisensi

Proyek ini dilisensikan di bawah Lisensi MIT. Lihat berkas `LICENSE` untuk detail lengkap.
