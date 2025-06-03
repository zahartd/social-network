-- ensure the DB exists
CREATE DATABASE IF NOT EXISTS stats;

-- raw events table
CREATE TABLE IF NOT EXISTS stats.events (
    event_date Date           DEFAULT toDate(event_time),
    event_time DateTime,
    user_id    String,
    entity_id  String,       -- post_id
    metric     String,       -- e.g. 'post-views', 'post-likes', 'post-unlikes', 'post-comments'
    cnt        UInt64
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_date)
ORDER BY (entity_id, metric, event_date);