package repositories

import (
	"context"
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
)

func (r *Repository) CreateEvent(
	ctx context.Context,
	eventPayload *dto.CreateEventRequest,
) (*models.Event, error){
	requestPayload := &models.Event{
		ID: utils.GenerateUUID(),
		EventID: eventPayload.EventID,
		EventType: eventPayload.EventType,
		Direction: eventPayload.Direction,
		Status: eventPayload.Status,
		Payload: eventPayload.Payload,
	}

	if err := r.DB.WithContext(ctx).Create(&requestPayload).Error; err != nil{
		return nil, err
	}

	var event *models.Event
	if errE := r.DB.WithContext(ctx).Where("event_id = ?", requestPayload.EventID).Take(&event).Error; errE != nil{
		return nil, errE
	}

	return event, nil
}

func(r *Repository) UpdateEventStatus(
	ctx context.Context,
	eventID uuid.UUID,
	status string,
) (*models.Event, error){
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	err := r.DB.WithContext(ctx).
		Model(&models.Event{}).
		Where("event_id = ?", eventID).
		Updates(updates).Error

	if err != nil {
		return nil, err
	}

	var event *models.Event
	if errE := r.DB.WithContext(ctx).Where("event_id = ?", eventID).Take(&event).Error; errE != nil{
		return nil, errE
	}

	return event, nil
}