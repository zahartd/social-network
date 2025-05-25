CREATE TABLE IF NOT EXISTS user_stats (
    user_id UUID,
    total_posts UInt64,
    total_likes UInt64,
    total_views UInt64,
    total_comments UInt64,
    updated_at DateTime
) ENGINE = AggregatingMergeTree()
PARTITION BY toYYYYMM(updated_at)
ORDER BY user_id;