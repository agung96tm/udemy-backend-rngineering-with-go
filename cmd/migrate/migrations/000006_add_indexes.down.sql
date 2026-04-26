-- DOWN Migration

-- Drop indexes if they exist
DROP INDEX IF EXISTS idx_comments_content;
DROP INDEX IF EXISTS idx_posts_title;
DROP INDEX IF EXISTS idx_posts_tags;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_posts_user_id;
DROP INDEX IF EXISTS idx_comments_post_id;

-- Drop extension if exists
DROP EXTENSION IF EXISTS pg_trgm;
