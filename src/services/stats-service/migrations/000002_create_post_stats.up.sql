CREATE TABLE IF NOT EXISTS post_stats (
    post_id UUID,
    likes_count UInt64,
    views_count UInt64,
    comments_count UInt64,
    updated_at DateTime
) ENGINE = AggregatingMergeTree()
PARTITION BY toYYYYMM(updated_at)
ORDER BY post_id;