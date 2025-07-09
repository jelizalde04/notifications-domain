# Sistema de Notificaciones en Tiempo Real

Sistema completo de notificaciones en Go que recibe datos por webhook y gRPC, los almacena en PostgreSQL y los envía en tiempo real al frontend mediante WebSocket.

## 🚀 Características

- **WebHook**: Recibe notificaciones desde sistemas externos
- **gRPC**: API para crear notificaciones desde microservicios
- **WebSocket**: Notificaciones en tiempo real al frontend
- **REST API**: Consultar y marcar notificaciones como leídas
- **PostgreSQL**: Almacenamiento persistente con conservación de estructura de datos
- **JWT**: Autenticación para WebSocket y endpoints REST

## 📋 Requisitos

- Go 1.19+
- PostgreSQL 12+
- Variables de entorno configuradas

## ⚙️ Variables de Entorno

Crear archivo `.env`:

```env
# Database Configuration
DB_HOST=your-postgres-host
DB_PORT=5432
DB_USER=your-username
DB_PASSWORD=your-password
NOTIFICATIONS_DB_NAME=notifications_db

# JWT Configuration
JWT_SECRET=your-secret-key
```

## 🗄️ Estructura de la Base de Datos

La tabla `notifications` se crea automáticamente con GORM:

```sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "actorId" UUID NOT NULL,
    "recipientId" UUID NOT NULL,
    "responsibleId" UUID NOT NULL,
    type VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    read BOOLEAN DEFAULT FALSE,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL
);
```

**Importante**: Los campos mantienen el formato camelCase original (`responsibleId`, `actorId`, etc.).

## 🏃‍♂️ Ejecución

1. **Instalar dependencias**:
```bash
go mod tidy
```

2. **Ejecutar el sistema**:
```bash
go run cmd/main.go
```

3. **Verificar el sistema** (opcional):
```bash
go run check_system.go
```

El sistema ejecutará:
- Servidor HTTP en puerto `8001`
- Servidor gRPC en puerto `50051`
- WebSocket en `/ws`

## 📡 API Endpoints

### Webhook (POST /webhook/like)
Recibe notificaciones desde sistemas externos:

```json
{
  "event": "like_created",
  "data": {
    "actorId": "550e8400-e29b-41d4-a716-446655440000",
    "recipientId": "550e8400-e29b-41d4-a716-446655440001",
    "responsibleId": "550e8400-e29b-41d4-a716-446655440002",
    "type": "like",
    "content": "John liked your post",
    "timestamp": "2024-01-15T10:30:00.123456"
  }
}
```

### WebSocket (GET /ws)
Conexión en tiempo real con autenticación JWT:

**Headers requeridos**:
```
Authorization: Bearer <jwt-token>
```

**Mensajes recibidos**:
```json
{
  "type": "notification",
  "id": "notification-uuid",
  "actorId": "actor-uuid",
  "recipientId": "recipient-uuid",
  "notificationType": "like",
  "content": "John liked your post",
  "timestamp": "2024-01-15T10:30:00Z",
  "read": false
}
```

### REST API

#### Obtener notificaciones (GET /notifications/{userId})
```bash
curl -H "Authorization: Bearer <jwt-token>" \
     "http://localhost:8001/notifications/{userId}?limit=20&unread=true"
```

Respuesta:
```json
{
  "notifications": [
    {
      "id": "notification-uuid",
      "actorId": "actor-uuid",
      "recipientId": "recipient-uuid",
      "responsibleId": "responsible-uuid",
      "type": "like",
      "content": "John liked your post",
      "read": false,
      "timestamp": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1
}
```

#### Marcar como leída (PUT /notifications/{notificationId}/read)
```bash
curl -X PUT \
     -H "Authorization: Bearer <jwt-token>" \
     "http://localhost:8001/notifications/{notificationId}/read"
```

## 🔌 gRPC

### Servicio: NotificationService
**Puerto**: 50051

#### Método: FollowCreated
```protobuf
rpc FollowCreated(FollowCreatedRequest) returns (NotificationResponse);

message FollowCreatedRequest {
  string actorId = 1;
  string recipientId = 2;
  string responsableId = 3;
  string type = 4;
  string content = 5;
}
```

## 🧪 Pruebas con Postman

### 1. Test WebSocket
1. **Nueva pestaña WebSocket** en Postman
2. **URL**: `ws://localhost:8001/ws`
3. **Headers**: `Authorization: Bearer <your-jwt-token>`
4. **Conectar** y esperar mensaje de bienvenida
5. **Verificar** que se reciben notificaciones pendientes

### 2. Test Webhook
```bash
POST http://localhost:8001/webhook/like
Content-Type: application/json

{
  "event": "like_created",
  "data": {
    "actorId": "550e8400-e29b-41d4-a716-446655440000",
    "recipientId": "550e8400-e29b-41d4-a716-446655440001", 
    "responsibleId": "USER_ID_FROM_JWT",
    "type": "like",
    "content": "Test notification",
    "timestamp": "2024-01-15T10:30:00.123456"
  }
}
```

### 3. Test REST API
```bash
# Obtener notificaciones
GET http://localhost:8001/notifications/{userId}
Authorization: Bearer <jwt-token>

# Marcar como leída
PUT http://localhost:8001/notifications/{notificationId}/read
Authorization: Bearer <jwt-token>
```

## 🔄 Flujo Completo

1. **Sistema externo** envía webhook o gRPC
2. **Notificación** se guarda en PostgreSQL con estructura original
3. **WebSocket** envía notificación en tiempo real si usuario conectado
4. **Usuario** consulta notificaciones via REST API
5. **Usuario** marca notificaciones como leídas

## 🛠️ Estructura del Proyecto

```
notifications/
├── cmd/main.go              # Punto de entrada
├── config/config.go         # Configuración DB y env
├── models/notification.go   # Modelo de datos
├── dto/notification.go      # DTOs para requests
├── handlers/               
│   ├── notification_handler.go  # REST API y webhook
│   └── ws_handler.go            # WebSocket
├── grpc/server.go          # Servidor gRPC
├── proto/                  # Archivos protobuf
├── utils/                  # JWT y utilidades
└── check_system.go         # Script de verificación
```

## 🔍 Logs y Debug

El sistema incluye logs detallados para:
- Conexiones WebSocket establecidas/cerradas
- Notificaciones enviadas por WebSocket
- Datos recibidos por webhook/gRPC
- Operaciones de base de datos
- Autenticación JWT

## ⚡ Características Técnicas

- **Sin Redis**: Conexiones WebSocket en memoria
- **Datos originales**: Se conserva estructura JSON exacta
- **Autenticación**: JWT para WebSocket y REST
- **Concurrencia**: Goroutines para gRPC y WebSocket
- **Error handling**: Logs detallados sin fallar operaciones
- **CORS**: Configurado para desarrollo

## 🐛 Troubleshooting

### Error "column does not exist"
Verificar que PostgreSQL tenga las columnas con nombres exactos:
```sql
SELECT column_name FROM information_schema.columns 
WHERE table_name = 'notifications';
```

### WebSocket no conecta
1. Verificar JWT válido en Authorization header
2. Confirmar formato: `Bearer <token>`
3. Revisar logs del servidor para errores de autenticación

### Notificaciones no llegan en tiempo real
1. Verificar usuario conectado a WebSocket
2. Confirmar `responsibleId` en webhook/gRPC coincide con `userId` del JWT
3. Revisar logs para verificar envío exitoso

---

**Estado**: ✅ Sistema completo y funcional  
**Última actualización**: 2024-01-15
