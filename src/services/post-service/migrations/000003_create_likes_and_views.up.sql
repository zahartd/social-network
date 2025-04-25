CREATE TABLE IF NOT EXISTS post_likes (
    user_id UUID NOT NULL,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY(user_id, post_id)
);
CREATE INDEX IF NOT EXISTS idx_post_likes_post_id ON post_likes(post_id);

CREATE TABLE IF NOT EXISTS post_views (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    viewed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_post_views_post_id ON post_views(post_id);