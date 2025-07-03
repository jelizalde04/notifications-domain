package dto

type LikeWebhookRequest struct {
	Event string          `json:"event" binding:"required"`
	Data  LikeWebhookData `json:"data" binding:"required"`
}

type LikeWebhookData struct {
	Type          string `json:"type" binding:"required"`
	ActorId       string `json:"actorId" binding:"required"`
	RecipientId   string `json:"recipientId" binding:"required"`
	ResponsibleId string `json:"responsibleId" binding:"required"`
	Timestamp     string `json:"timestamp" binding:"required"`
	Content       string `json:"content" binding:"required"`
}
