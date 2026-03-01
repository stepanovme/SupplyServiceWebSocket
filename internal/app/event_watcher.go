package app

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"supplyservicews/internal/db"
	"supplyservicews/internal/ws"
)

type EventWatcher struct {
	repo         *db.Repository
	hub          *ws.Hub
	pollInterval time.Duration
	lastEventID  int64
}

func NewEventWatcher(repo *db.Repository, hub *ws.Hub, pollInterval time.Duration) *EventWatcher {
	return &EventWatcher{
		repo:         repo,
		hub:          hub,
		pollInterval: pollInterval,
	}
}

func (w *EventWatcher) Init(ctx context.Context) error {
	if err := w.repo.EnsureEventInfrastructure(ctx); err != nil {
		return err
	}

	lastID, err := w.repo.LastEventID(ctx)
	if err != nil {
		return err
	}
	w.lastEventID = lastID

	return nil
}

func (w *EventWatcher) Run(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.processBatch(ctx); err != nil {
				log.Printf("event watcher error: %v", err)
			}
		}
	}
}

func (w *EventWatcher) processBatch(ctx context.Context) error {
	events, err := w.repo.FetchEventsAfter(ctx, w.lastEventID, 100)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}

	for _, event := range events {
		if err := w.dispatchEvent(ctx, event); err != nil {
			log.Printf("dispatch event %d failed: %v", event.ID, err)
		}
		w.lastEventID = event.ID
	}

	return nil
}

func (w *EventWatcher) dispatchEvent(ctx context.Context, event db.DBEvent) error {
	if event.EventType == "request_deleted" {
		message := map[string]any{
			"type":          event.EventType,
			"request_id":    event.RequestID,
			"source_table":  event.SourceTable,
			"source_action": event.SourceAction,
			"event_id":      event.ID,
			"created_at":    event.CreatedAt,
			"deleted_data":  event.Payload,
		}
		return w.hub.Broadcast(ctx, message)
	}

	snapshot, err := w.repo.GetRequestSnapshot(ctx, event.RequestID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	message := map[string]any{
		"type":          event.EventType,
		"request_id":    event.RequestID,
		"source_table":  event.SourceTable,
		"source_action": event.SourceAction,
		"event_id":      event.ID,
		"created_at":    event.CreatedAt,
		"data":          snapshot,
	}

	if event.EventType == "request_notification" {
		notification, err := db.ParseNotificationPayload(event.Payload)
		if err != nil {
			return err
		}
		message["notification"] = notification
	}

	return w.hub.Broadcast(ctx, message)
}
