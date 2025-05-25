CREATE TABLE IF NOT EXISTS event_log (
    event_id UUID,
    event_type String,
    user_id UUID,
    target_id UUID,
    timestamp DateTime,
    details JSON
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (event_type, target_id, timestamp);