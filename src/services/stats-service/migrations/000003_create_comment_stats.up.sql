CREATE TABLE IF NOT EXISTS comment_stats (
    comment_id UUID,
    likes_count UInt64,
    reply_count UInt64,
    updated_at DateTime
) ENGINE = AggregatingMergeTree()
PARTITION BY toYYYYMM(updated_at)
ORDER BY comment_id;