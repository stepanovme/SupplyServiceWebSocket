package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) EnsureEventInfrastructure(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS request_events (
			id BIGINT NOT NULL AUTO_INCREMENT,
			request_id INT NOT NULL,
			event_type ENUM('request_created', 'request_updated', 'request_deleted', 'request_notification') NOT NULL,
			source_table VARCHAR(64) NOT NULL,
			source_action ENUM('insert', 'update', 'delete') NOT NULL,
			payload JSON NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			KEY idx_request_id (request_id),
			KEY idx_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci`,
		`ALTER TABLE request_events
		MODIFY COLUMN event_type ENUM('request_created', 'request_updated', 'request_deleted', 'request_notification') NOT NULL`,
		`ALTER TABLE request_events
		MODIFY COLUMN source_action ENUM('insert', 'update', 'delete') NOT NULL`,
	}

	for _, query := range queries {
		if _, err := r.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("ensure request_events table: %w", err)
		}
	}

	triggers := []triggerDefinition{
		{
			Name: "trg_request_ai_ws",
			SQL: `CREATE TRIGGER trg_request_ai_ws AFTER INSERT ON request
			FOR EACH ROW
			INSERT INTO request_events (request_id, event_type, source_table, source_action)
			VALUES (NEW.id, 'request_created', 'request', 'insert')`,
		},
		{
			Name: "trg_request_au_ws",
			SQL: `CREATE TRIGGER trg_request_au_ws AFTER UPDATE ON request
			FOR EACH ROW
			INSERT INTO request_events (request_id, event_type, source_table, source_action)
			VALUES (NEW.id, 'request_updated', 'request', 'update')`,
		},
		{
			Name: "trg_request_ad_ws",
			SQL: `CREATE TRIGGER trg_request_ad_ws AFTER DELETE ON request
			FOR EACH ROW
			INSERT INTO request_events (request_id, event_type, source_table, source_action, payload)
			VALUES (
				OLD.id,
				'request_deleted',
				'request',
				'delete',
				JSON_OBJECT(
					'id', OLD.id,
					'object_levels_id', OLD.object_levels_id,
					'name', OLD.name,
					'created_by', OLD.created_by,
					'executor', OLD.executor,
					'status_id', OLD.status_id
				)
			)`,
		},
		{
			Name: "trg_request_items_ai_ws",
			SQL: `CREATE TRIGGER trg_request_items_ai_ws AFTER INSERT ON request_items
			FOR EACH ROW
			INSERT INTO request_events (request_id, event_type, source_table, source_action)
			VALUES (NEW.request_id, 'request_updated', 'request_items', 'insert')`,
		},
		{
			Name: "trg_request_items_au_ws",
			SQL: `CREATE TRIGGER trg_request_items_au_ws AFTER UPDATE ON request_items
			FOR EACH ROW
			INSERT INTO request_events (request_id, event_type, source_table, source_action)
			VALUES (NEW.request_id, 'request_updated', 'request_items', 'update')`,
		},
		{
			Name: "trg_request_items_ad_ws",
			SQL: `CREATE TRIGGER trg_request_items_ad_ws AFTER DELETE ON request_items
			FOR EACH ROW
			INSERT INTO request_events (request_id, event_type, source_table, source_action)
			VALUES (OLD.request_id, 'request_updated', 'request_items', 'delete')`,
		},
		{
			Name: "trg_request_files_ai_ws",
			SQL: `CREATE TRIGGER trg_request_files_ai_ws AFTER INSERT ON request_files
			FOR EACH ROW
			INSERT INTO request_events (request_id, event_type, source_table, source_action)
			VALUES (NEW.request_id, 'request_updated', 'request_files', 'insert')`,
		},
		{
			Name: "trg_request_files_au_ws",
			SQL: `CREATE TRIGGER trg_request_files_au_ws AFTER UPDATE ON request_files
			FOR EACH ROW
			INSERT INTO request_events (request_id, event_type, source_table, source_action)
			VALUES (NEW.request_id, 'request_updated', 'request_files', 'update')`,
		},
		{
			Name: "trg_request_files_ad_ws",
			SQL: `CREATE TRIGGER trg_request_files_ad_ws AFTER DELETE ON request_files
			FOR EACH ROW
			INSERT INTO request_events (request_id, event_type, source_table, source_action)
			VALUES (OLD.request_id, 'request_updated', 'request_files', 'delete')`,
		},
		{
			Name: "trg_request_log_ai_ws",
			SQL: `CREATE TRIGGER trg_request_log_ai_ws AFTER INSERT ON request_log
			FOR EACH ROW
			INSERT INTO request_events (request_id, event_type, source_table, source_action, payload)
			VALUES (
				CAST(NEW.request_id AS UNSIGNED),
				'request_notification',
				'request_log',
				'insert',
				JSON_OBJECT(
					'log_id', NEW.id,
					'user_id', NEW.user_id,
					'status_name', NEW.status_name,
					'date_response', DATE_FORMAT(NEW.date_response, '%Y-%m-%d %H:%i:%s'),
					'request_log_request_id', NEW.request_id
				)
			)`,
		},
		{
			Name: "trg_request_log_au_ws",
			SQL: `CREATE TRIGGER trg_request_log_au_ws AFTER UPDATE ON request_log
			FOR EACH ROW
			INSERT INTO request_events (request_id, event_type, source_table, source_action, payload)
			VALUES (
				CAST(NEW.request_id AS UNSIGNED),
				'request_notification',
				'request_log',
				'update',
				JSON_OBJECT(
					'log_id', NEW.id,
					'user_id', NEW.user_id,
					'status_name', NEW.status_name,
					'date_response', DATE_FORMAT(NEW.date_response, '%Y-%m-%d %H:%i:%s'),
					'request_log_request_id', NEW.request_id
				)
			)`,
		},
		{
			Name: "trg_request_log_ad_ws",
			SQL: `CREATE TRIGGER trg_request_log_ad_ws AFTER DELETE ON request_log
			FOR EACH ROW
			INSERT INTO request_events (request_id, event_type, source_table, source_action)
			VALUES (CAST(OLD.request_id AS UNSIGNED), 'request_updated', 'request_log', 'delete')`,
		},
	}

	for _, trg := range triggers {
		exists, err := r.triggerExists(ctx, trg.Name)
		if err != nil {
			return fmt.Errorf("check trigger %s: %w", trg.Name, err)
		}
		if exists {
			continue
		}

		if _, err := r.db.ExecContext(ctx, trg.SQL); err != nil {
			return fmt.Errorf("create trigger %s: %w", trg.Name, err)
		}
	}

	return nil
}

func (r *Repository) triggerExists(ctx context.Context, name string) (bool, error) {
	const query = `SELECT COUNT(*)
	FROM information_schema.TRIGGERS
	WHERE TRIGGER_SCHEMA = DATABASE() AND TRIGGER_NAME = ?`

	var count int
	if err := r.db.QueryRowContext(ctx, query, name).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) LastEventID(ctx context.Context) (int64, error) {
	const query = `SELECT COALESCE(MAX(id), 0) FROM request_events`
	var id int64
	if err := r.db.QueryRowContext(ctx, query).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repository) FetchEventsAfter(ctx context.Context, lastID int64, limit int) ([]DBEvent, error) {
	const query = `SELECT id, request_id, event_type, source_table, source_action, COALESCE(payload, JSON_OBJECT()), created_at
	FROM request_events
	WHERE id > ?
	ORDER BY id ASC
	LIMIT ?`

	rows, err := r.db.QueryContext(ctx, query, lastID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]DBEvent, 0, limit)
	for rows.Next() {
		var event DBEvent
		var payloadBytes []byte

		if err := rows.Scan(&event.ID, &event.RequestID, &event.EventType, &event.SourceTable, &event.SourceAction, &payloadBytes, &event.CreatedAt); err != nil {
			return nil, err
		}
		event.Payload = append(event.Payload[:0], payloadBytes...)
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *Repository) GetRequestSnapshot(ctx context.Context, requestID int) (RequestSnapshot, error) {
	snapshot := RequestSnapshot{}

	if err := r.getRequest(ctx, requestID, &snapshot.Request); err != nil {
		return RequestSnapshot{}, err
	}

	items, err := r.getRequestItems(ctx, requestID)
	if err != nil {
		return RequestSnapshot{}, err
	}
	snapshot.Items = items

	files, err := r.getRequestFiles(ctx, requestID)
	if err != nil {
		return RequestSnapshot{}, err
	}
	snapshot.Files = files

	logs, err := r.getRequestLogs(ctx, requestID)
	if err != nil {
		return RequestSnapshot{}, err
	}
	snapshot.Logs = logs

	return snapshot, nil
}

func (r *Repository) getRequest(ctx context.Context, requestID int, request *Request) error {
	const query = `SELECT id, object_levels_id, name, created_by, executor, created_at, started_at, approved_at, rejected_at, completed_at, deadline, comment, status_id
	FROM request
	WHERE id = ?`

	if err := r.db.QueryRowContext(ctx, query, requestID).Scan(
		&request.ID,
		&request.ObjectLevelsID,
		&request.Name,
		&request.CreatedBy,
		&request.Executor,
		&request.CreatedAt,
		&request.StartedAt,
		&request.ApprovedAt,
		&request.RejectedAt,
		&request.CompletedAt,
		&request.Deadline,
		&request.Comment,
		&request.StatusID,
	); err != nil {
		return err
	}

	return nil
}

func (r *Repository) getRequestItems(ctx context.Context, requestID int) ([]RequestItem, error) {
	const query = `SELECT id, request_id, num, nomenclature_id, name, unit_id, quantity, warehouse_category_id, comment
	FROM request_items
	WHERE request_id = ?
	ORDER BY num ASC`

	rows, err := r.db.QueryContext(ctx, query, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]RequestItem, 0)
	for rows.Next() {
		var item RequestItem
		if err := rows.Scan(
			&item.ID,
			&item.RequestID,
			&item.Num,
			&item.NomenclatureID,
			&item.Name,
			&item.UnitID,
			&item.Quantity,
			&item.WarehouseCategoryID,
			&item.Comment,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *Repository) getRequestFiles(ctx context.Context, requestID int) ([]RequestFile, error) {
	const query = `SELECT id, request_id, file_id, link_type, description, is_main, sort_order, created_at, created_by
	FROM request_files
	WHERE request_id = ?
	ORDER BY sort_order ASC, created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := make([]RequestFile, 0)
	for rows.Next() {
		var f RequestFile
		if err := rows.Scan(
			&f.ID,
			&f.RequestID,
			&f.FileID,
			&f.LinkType,
			&f.Description,
			&f.IsMain,
			&f.SortOrder,
			&f.CreatedAt,
			&f.CreatedBy,
		); err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

func (r *Repository) getRequestLogs(ctx context.Context, requestID int) ([]RequestLog, error) {
	const query = `SELECT id, user_id, request_id, status_name, date_response
	FROM request_log
	WHERE CAST(request_id AS UNSIGNED) = ?
	ORDER BY id ASC`

	rows, err := r.db.QueryContext(ctx, query, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]RequestLog, 0)
	for rows.Next() {
		var l RequestLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.RequestID, &l.StatusName, &l.DateResponse); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

func ParseNotificationPayload(payload json.RawMessage) (NotificationPayload, error) {
	var p NotificationPayload
	if len(payload) == 0 {
		return p, nil
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return p, err
	}
	return p, nil
}

type triggerDefinition struct {
	Name string
	SQL  string
}
