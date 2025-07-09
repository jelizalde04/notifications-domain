package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"notifications/config"
	"notifications/handlers"
	"notifications/models"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"gorm.io/gorm/clause"

	pb "notifications/proto/notificationpb"
)

type NotificationGRPCServer struct {
	pb.UnimplementedNotificationServiceServer
}

// Método que maneja la llamada FollowCreated
func (s *NotificationGRPCServer) FollowCreated(ctx context.Context, req *pb.FollowCreatedRequest) (*pb.NotificationResponse, error) {
	actorID, err := uuid.Parse(req.GetActorId())
	if err != nil {
		return nil, fmt.Errorf("invalid actorId: %w", err)
	}

	recipientID, err := uuid.Parse(req.GetRecipientId())
	if err != nil {
		return nil, fmt.Errorf("invalid recipientId: %w", err)
	}

	responsibleID, err := uuid.Parse(req.GetResponsableId())
	if err != nil {
		return nil, fmt.Errorf("invalid responsableId: %w", err)
	}

	noti := models.Notification{
		ActorID:       actorID,
		RecipientID:   recipientID,
		ResponsibleID: responsibleID,
		Type:          req.GetType(),
		Content:       req.GetContent(),
		Timestamp:     time.Now(),
	}

	if err := config.DB.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "actorId"}, {Name: "recipientId"}, {Name: "type"}, {Name: "content"}},
			DoUpdates: clause.AssignmentColumns([]string{"timestamp"}),
		}).
		Create(&noti).Error; err != nil {
		return nil, fmt.Errorf("failed to save notification: %w", err)
	}

	log.Println("Notification saved to DB:", noti.ID.String())

	// Enviar notificación por WebSocket si el usuario está conectado
	notificationMessage := fmt.Sprintf(`{
		"type": "notification",
		"id": "%s",
		"actorId": "%s",
		"recipientId": "%s",
		"notificationType": "%s",
		"content": "%s",
		"timestamp": "%s",
		"read": false
	}`, noti.ID, noti.ActorID, noti.RecipientID,
		noti.Type, noti.Content, noti.Timestamp.Format(time.RFC3339))

	// Intentar enviar por WebSocket (no falla si el usuario no está conectado)
	log.Printf("Attempting to send WebSocket notification to responsibleId: %s", req.GetResponsableId())
	if err := handlers.SendNotification(req.GetResponsableId(), notificationMessage); err != nil {
		// Log pero no falla la operación
		log.Printf("Could not send WebSocket notification to user %s: %v", req.GetResponsableId(), err)
	} else {
		log.Printf("WebSocket notification sent successfully to user %s", req.GetResponsableId())
	}

	return &pb.NotificationResponse{
		Message: "Notification saved successfully",
	}, nil
}

// StartGRPCServer arranca el servidor gRPC
func StartGRPCServer() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterNotificationServiceServer(server, &NotificationGRPCServer{})

	log.Println("gRPC server listening on :50051")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}
