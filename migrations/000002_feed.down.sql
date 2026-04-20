DROP TABLE IF EXISTS post_comments;
DROP TABLE IF EXISTS post_likes;
DROP TABLE IF EXISTS posts;

DROP FUNCTION IF EXISTS inc_post_like_count();
DROP FUNCTION IF EXISTS dec_post_like_count();
DROP FUNCTION IF EXISTS inc_post_comment_count();
DROP FUNCTION IF EXISTS dec_post_comment_count();
