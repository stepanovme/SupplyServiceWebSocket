CREATE TABLE IF NOT EXISTS request_events (
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

ALTER TABLE request_events
    MODIFY COLUMN event_type ENUM('request_created', 'request_updated', 'request_deleted', 'request_notification') NOT NULL;

ALTER TABLE request_events
    MODIFY COLUMN source_action ENUM('insert', 'update', 'delete') NOT NULL;

DELIMITER $$

CREATE TRIGGER trg_request_ai_ws
AFTER INSERT ON request
FOR EACH ROW
BEGIN
    INSERT INTO request_events (request_id, event_type, source_table, source_action)
    VALUES (NEW.id, 'request_created', 'request', 'insert');
END$$

CREATE TRIGGER trg_request_au_ws
AFTER UPDATE ON request
FOR EACH ROW
BEGIN
    INSERT INTO request_events (request_id, event_type, source_table, source_action)
    VALUES (NEW.id, 'request_updated', 'request', 'update');
END$$

CREATE TRIGGER trg_request_ad_ws
AFTER DELETE ON request
FOR EACH ROW
BEGIN
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
    );
END$$

CREATE TRIGGER trg_request_items_ai_ws
AFTER INSERT ON request_items
FOR EACH ROW
BEGIN
    INSERT INTO request_events (request_id, event_type, source_table, source_action)
    VALUES (NEW.request_id, 'request_updated', 'request_items', 'insert');
END$$

CREATE TRIGGER trg_request_items_au_ws
AFTER UPDATE ON request_items
FOR EACH ROW
BEGIN
    INSERT INTO request_events (request_id, event_type, source_table, source_action)
    VALUES (NEW.request_id, 'request_updated', 'request_items', 'update');
END$$

CREATE TRIGGER trg_request_items_ad_ws
AFTER DELETE ON request_items
FOR EACH ROW
BEGIN
    INSERT INTO request_events (request_id, event_type, source_table, source_action)
    VALUES (OLD.request_id, 'request_updated', 'request_items', 'delete');
END$$

CREATE TRIGGER trg_request_files_ai_ws
AFTER INSERT ON request_files
FOR EACH ROW
BEGIN
    INSERT INTO request_events (request_id, event_type, source_table, source_action)
    VALUES (NEW.request_id, 'request_updated', 'request_files', 'insert');
END$$

CREATE TRIGGER trg_request_files_au_ws
AFTER UPDATE ON request_files
FOR EACH ROW
BEGIN
    INSERT INTO request_events (request_id, event_type, source_table, source_action)
    VALUES (NEW.request_id, 'request_updated', 'request_files', 'update');
END$$

CREATE TRIGGER trg_request_files_ad_ws
AFTER DELETE ON request_files
FOR EACH ROW
BEGIN
    INSERT INTO request_events (request_id, event_type, source_table, source_action)
    VALUES (OLD.request_id, 'request_updated', 'request_files', 'delete');
END$$

CREATE TRIGGER trg_request_log_ai_ws
AFTER INSERT ON request_log
FOR EACH ROW
BEGIN
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
    );
END$$

CREATE TRIGGER trg_request_log_au_ws
AFTER UPDATE ON request_log
FOR EACH ROW
BEGIN
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
    );
END$$

CREATE TRIGGER trg_request_log_ad_ws
AFTER DELETE ON request_log
FOR EACH ROW
BEGIN
    INSERT INTO request_events (request_id, event_type, source_table, source_action)
    VALUES (CAST(OLD.request_id AS UNSIGNED), 'request_updated', 'request_log', 'delete');
END$$

DELIMITER ;
