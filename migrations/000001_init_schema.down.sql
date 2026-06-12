-- 000001_init_schema.down.sql

DROP INDEX IF EXISTS idx_comment_reactions_comment_id;
DROP INDEX IF EXISTS idx_post_reactions_post_id;
DROP INDEX IF EXISTS idx_comments_user_id;
DROP INDEX IF EXISTS idx_comments_post_id;
DROP INDEX IF EXISTS idx_posts_category_id;
DROP INDEX IF EXISTS idx_posts_author_id;

DROP TABLE IF EXISTS comment_reactions;
DROP TABLE IF EXISTS post_reactions;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS users;