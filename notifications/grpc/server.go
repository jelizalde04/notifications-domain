package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"notifications/config"
	"notifications/models"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"

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

	recipientID, err := uuid.Parse(req.GetResponsableId())
	if err != nil {
		return nil, fmt.Errorf("invalid responsableId: %w", err)
	}

	noti := models.Notification{
		ActorID:     actorID,
		RecipientID: recipientID,
		Type:        req.GetType(),
		Content:     req.GetContent(),
		Timestamp:   time.Now(),
	}

	if err := config.DB.Create(&noti).Error; err != nil {
		return nil, fmt.Errorf("failed to save notification: %w", err)
	}

	log.Println("Notification saved to DB:", noti.ID.String())

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
