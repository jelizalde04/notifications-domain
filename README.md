# Notifications Domain 🔔

Notification system developed in Go that provides comprehensive real-time notification functionalities through WebSockets, REST API, and gRPC.

## 🏗️ General Architecture
   
This project implements a notification microservice with the following features:

- **REST API** with Gin for CRUD operations
- **WebSockets** for real-time notifications
- **gRPC** for inter-microservice communication
- **PostgreSQL** as the main database
- **JWT** for authentication
- **Docker** for containerization

## 📁 Project Structure

```
notifications/
├── cmd/                    # Application entry point
│   └── main.go            # Main server (HTTP + gRPC)
├── config/                # Application configuration
│   └── config.go          # DB connection and environment variables
├── dto/                   # Data Transfer Objects
│   └── notification.go    # DTOs for webhooks and requests
├── grpc/                  # gRPC server
│   └── server.go          # gRPC server implementation
├── handlers/              # HTTP controllers
│   ├── notification_handler.go  # REST endpoints
│   └── ws_handler.go      # WebSocket handler
├── internal/              # Internal code
│   └── websocket/
│       └── server.go      # WebSocket server
├── models/                # Data models
│   └── notification.go    # Notification model (GORM)
├── proto/                 # Protocol Buffers definitions
│   ├── notification.proto # gRPC service definition
│   └── notificationpb/    # Generated code
│       ├── notification_grpc.pb.go
│       └── notification.pb.go
├── utils/                 # Utilities
│   ├── broadcast.go       # Broadcasting system
│   └── jwt.go            # JWT handling
├── go.mod                # Go dependencies
├── go.sum                # Dependency checksums
├── Dockerfile            # Docker image
├── docker-compose.yml    # Orchestration with PostgreSQL
└── README.md            # This documentation
```

## 🚀 Features

### 1. REST API (Port 8001)
- `GET /ping` - Health check
- `POST /webhook/like` - Webhook to process likes
- `GET /notifications/:userId` - Get user notifications
- `PUT /notifications/:notificationId/read` - Mark notification as read
- `GET /ws` - Upgrade to WebSocket connection

### 2. gRPC Service (Port 9001)
- `FollowCreated` - Create new follower notification

### 3. WebSockets
- Real-time notifications
- Broadcasting to specific users
- JWT authentication

## 🛠️ Technologies Used

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Web Framework** | Gin | REST API and middleware |
| **WebSockets** | Gorilla WebSocket | Real-time communication |
| **gRPC** | Google gRPC | Inter-service communication |
| **ORM** | GORM | Object-relational mapping |
| **Database** | PostgreSQL | Data persistence |
| **Authentication** | JWT | Access tokens |
| **Containerization** | Docker | Deployment and development |
| **UUID** | Google UUID | Unique identifiers |
| **Configuration** | GoDotEnv | Environment variables |

## 📊 Data Model

### Notification
```go
type Notification struct {
    ID            uuid.UUID  // Unique identifier
    ActorID       uuid.UUID  // User who performs the action
    RecipientID   uuid.UUID  // User who receives the notification
    ResponsibleID uuid.UUID  // User responsible for the content
    Type          string     // Notification type (like, follow, etc.)
    Content       string     // Descriptive content
    Read          bool       // Read status
    Timestamp     time.Time  // Creation time
}
```

## 🔧 Configuration

### Required Environment Variables
```bash
DB_HOST=localhost           # PostgreSQL host
DB_PORT=5432               # PostgreSQL port
DB_USER=notifications_user # Database user
DB_PASSWORD=password       # Database password
NOTIFICATIONS_DB_NAME=db   # Database name
```

## 🐳 Docker Deployment

### Local Development
```bash
# Clone the repository
git clone <repository-url>
cd notifications-domain/notifications

# Configure environment variables
cp .env.example .env

# Run with Docker Compose (includes PostgreSQL)
docker-compose up --build

# The application will be available at:
# - HTTP API: http://localhost:8001
# - WebSocket: ws://localhost:8001/ws
# - PostgreSQL: localhost:5432
```

### Application Only
```bash
# Build image
docker build -t notifications-service .

# Run container
docker run -d \
  --name notifications-app \
  -p 8001:8001 \
  -p 9001:9001 \
  --env-file .env \
  notifications-service
```

## 📡 Data Flow

1. **Webhook Reception**: External events arrive via POST `/webhook/like`
2. **Processing**: Information is validated and processed
3. **Database Storage**: Persisted in PostgreSQL using GORM
4. **Real-time Broadcast**: Sent via WebSocket to the target user
5. **gRPC Communication**: Other services can create notifications via gRPC

## 🔐 Security

- **CORS** configured for development
- **JWT Authentication** for WebSockets
- **Data validation** on all endpoints
- **Non-root user** in Docker containers
- **Health checks** for monitoring

## 📈 Scalability

- **Stateless design** for horizontal scaling
- **gRPC** for efficient inter-service communication
- **WebSockets with broadcasting** for real-time notifications
- **Containerization** for Kubernetes deployment
- **Database pooling** via GORM

## 🧪 Testing

```bash
# Ejecutar tests
go test ./...

# Test de cobertura
go test -cover ./...

# Health check
curl http://localhost:8001/ping
```
