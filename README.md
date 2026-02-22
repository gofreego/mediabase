# Mediabase

A generic media storage service built with Go, supporting presigned uploads and downloads using MinIO (S3-compatible storage). The service is designed with interface abstractions to allow easy migration to AWS S3 or other storage providers.

## Features

- **Presigned Upload Policies**: Generate secure, time-limited URLs and form policies for direct file uploads. Enforces constraints strictly on the server/storage side.
- **Presigned Download URLs**: Generate secure, time-limited URLs for file downloads.
- **Bucket Creation & Management**: Seamlessly configure buckets programmatically, with optional public-read permissions.
- **Granular Paths & Filenames**: Organize files dynamically with requested paths and specific filenames, or let the service generate UUIDs automatically.
- **Object Management**: Delete files directly via API.
- **Storage Abstraction**: Interface-based design for easy migration between storage providers (MinIO, S3, GCS, etc.).
- **gRPC Gateway**: HTTP/REST API automatically generated from Protocol Buffers.
- **Swagger Documentation**: Auto-generated API documentation.
- **Interactive Test Console**: Detailed web console included (`test/test.html`) to test all functionalities.

## Architecture

The service follows a clean architecture pattern:

```text
├── api/                    # Protocol Buffer definitions and generated code
│   ├── proto/              # .proto files
│   └── mediabase_v1/       # Generated Go code
├── cmd/                    # Application entry points
│   ├── http_server/        # HTTP server implementation
│   └── grpc_server/        # gRPC server implementation
├── internal/
│   ├── configs/            # Configuration management
│   ├── repository/         # Data access layer
│   ├── service/            # Business logic
│   └── storage/            # Storage abstraction
│       └── minio/          # MinIO implementation
├── resources/              # Setup scripts
├── test/                   # Interactive Test Console (test.html)
└── main.go                 # Application entry point
```

## Prerequisites

- Go 1.23+
- Docker & Docker Compose (for running MinIO)
- Protocol Buffer compiler and plugins (installed via `make install`)

## Installation

1. **Clone the repository**
```bash
git clone <repository-url>
cd mediabase
```

2. **Install dependencies**
```bash
make install
```

3. **Generate Protocol Buffer code**
```bash
make setup
```

## Running the Service

### 1. Start MinIO (Object Storage)

```bash
docker-compose up -d
```

This will start MinIO on:
- **API**: http://localhost:9000
- **Console**: http://localhost:9001
- **Credentials**: minioadmin/minioadmin

### 2. Start the Mediabase Service

```bash
make run
# OR
go run main.go
```

The service will start on:
- **HTTP Server**: http://localhost:8085
- **gRPC Server**: http://localhost:8086
- **Swagger UI**: http://localhost:8085/mediabase/v1/swagger

## API Endpoints

### 1. Create Bucket
Creates a bucket in storage and can optionally configure the bucket policy to allow public reads for downloads while keeping uploads strictly restricted.

**POST** `/api/upload/bucket`

Request:
```json
{
  "bucket_name": "mediatest",
  "is_public": true
}
```

Response:
```json
{
  "success": true
}
```

### 2. Generate Presigned Upload Policy
Returns a policy for secure uploads, allowing storage-level enforcement for file sizes and preventing unauthorized uploads.

**POST** `/api/upload/presign/upload`

Request:
```json
{
  "bucket_name": "mediatest",
  "content_type": "image/jpeg",
  "max_file_size": 5242880,
  "path": "users/avatars",
  "file_name": "avatar.jpg"
}
```

Response:
```json
{
  "presigned_url": "http://localhost:9000/mediatest",
  "object_key": "users/avatars/avatar.jpg",
  "expires_in": 60,
  "form_data": {
    "bucket": "mediatest",
    "key": "users/avatars/avatar.jpg",
    "policy": "eyJleHBpcmF0aW9uIjoi...",
    "x-amz-algorithm": "AWS4-HMAC-SHA256",
    "x-amz-credential": "...",
    "x-amz-date": "...",
    "x-amz-signature": "..."
  }
}
```

### 3. Generate Presigned Download URL

**POST** `/api/upload/presign/download`

Request:
```json
{
  "bucket_name": "mediatest",
  "object_key": "users/avatars/avatar.jpg"
}
```

Response:
```json
{
  "presigned_url": "http://localhost:9000/mediatest/users/avatars/avatar.jpg?X-Amz-...",
  "expires_in": 3600
}
```

### 4. Delete Object

**DELETE** `/api/upload/object/{object_key}?bucket_name={bucket_name}`

Response:
```json
{
  "success": true
}
```

## Configuration

Configuration is managed through YAML files. See `dev.yaml` for an example.

```yaml
Storage:
  Endpoint: "localhost:9000"
  AccessKeyID: "minioadmin"
  SecretAccessKey: "minioadmin"
  Region: "us-east-1"
  UseSSL: false
```

### Environment-specific Configurations

- `dev.yaml` - Development environment
- `prod.yaml` - Production environment (create as needed)

## Storage Provider Migration

The service uses an interface-based storage abstraction, making it easy to switch between providers:

### Current: MinIO
```go
storage, err := minio.NewMinIOStorage(config.Storage)
```

### Future: AWS S3
```go
storage, err := s3.NewS3Storage(config.Storage)
```

To add a new storage provider:
1. Implement the `storage.Storage` interface inside the `internal/storage` section.
2. Update the initialization in `cmd/http_server/http.go` and `cmd/grpc_server/grpc.go`.

## Interactive Test Console
A rich web-based interaction page is provided to visualize the granular 2-step upload sequence directly against MinIO. 
1. Run the APIs as listed above.
2. Open `test/test.html` in any modern web browser. 
3. Perform end-to-end tests validating policies, constraints, form data, fetching, and object deletion.

## File Upload Flow

1. **Client requests presigned upload policy**
   - POST `/api/upload/presign/upload` with `bucket_name`, `content_type`, `max_file_size`, and optional `path` and `file_name`.
   - Service generates the respective `object_key` and POST upload form policy (`form_data`).
2. **Client uploads file directly to MinIO**
   - Client executes an HTTP POST to `presigned_url` (bucket endpoint) using multi-part `FormData`.
   - Append all keys from `form_data` into Form Data, followed by appending `file` containing actual content last.
   - Storage enforces file-type & dimension controls. No server involvement in actual network transfer.
3. **Client can request presigned download URL**
   - POST `/api/upload/presign/download` indicating `bucket_name` + `object_key`.
   - Temporary signed URL generated.
4. **Client downloads file directly from MinIO**
   - GET from `presigned_url`.

## Constraints

- Constraints such as **Max file size** and **Allowed content types** depend on policy logic and validations you set in the service config or via UI.
- **Upload URL expiry**: 60 seconds (defaults).
- **Download URL expiry**: 3600 seconds (1 hour) defaults.

## License

See [LICENSE](LICENSE) file for details.
