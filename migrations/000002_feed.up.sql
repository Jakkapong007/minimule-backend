-- miniMule feed: posts, likes, comments
-- Run: make migrate-up

-- ── Posts ─────────────────────────────────────────────────────────────────────
CREATE TABLE posts (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    caption           TEXT,
    image_url         TEXT        NOT NULL,
    is_sticker_design BOOLEAN     NOT NULL DEFAULT FALSE,
    like_count        INT         NOT NULL DEFAULT 0,
    comment_count     INT         NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ
);

CREATE INDEX idx_posts_user       ON posts (user_id);
CREATE INDEX idx_posts_created_at ON posts (created_at DESC);
CREATE INDEX idx_posts_sticker    ON posts (is_sticker_design) WHERE is_sticker_design = TRUE;

-- ── Post Likes ────────────────────────────────────────────────────────────────
CREATE TABLE post_likes (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id    UUID        NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (post_id, user_id)
);

CREATE INDEX idx_post_likes_post ON post_likes (post_id);
CREATE INDEX idx_post_likes_user ON post_likes (user_id);

-- ── Post Comments ─────────────────────────────────────────────────────────────
CREATE TABLE post_comments (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id    UUID        NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_post_comments_post ON post_comments (post_id);
CREATE INDEX idx_post_comments_user ON post_comments (user_id);

-- ── Denorm triggers for like/comment counts ───────────────────────────────────
CREATE OR REPLACE FUNCTION inc_post_like_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE posts SET like_count = like_count + 1 WHERE id = NEW.post_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION dec_post_like_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE posts SET like_count = GREATEST(like_count - 1, 0) WHERE id = OLD.post_id;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION inc_post_comment_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE posts SET comment_count = comment_count + 1 WHERE id = NEW.post_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION dec_post_comment_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE posts SET comment_count = GREATEST(comment_count - 1, 0) WHERE id = OLD.post_id;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_post_likes_insert
    AFTER INSERT ON post_likes FOR EACH ROW EXECUTE FUNCTION inc_post_like_count();

CREATE TRIGGER trg_post_likes_delete
    AFTER DELETE ON post_likes FOR EACH ROW EXECUTE FUNCTION dec_post_like_count();

CREATE TRIGGER trg_post_comments_insert
    AFTER INSERT ON post_comments FOR EACH ROW EXECUTE FUNCTION inc_post_comment_count();

CREATE TRIGGER trg_post_comments_delete
    AFTER DELETE ON post_comments FOR EACH ROW EXECUTE FUNCTION dec_post_comment_count();

CREATE TRIGGER trg_posts_updated_at
    BEFORE UPDATE ON posts FOR EACH ROW EXECUTE FUNCTION set_updated_at();
