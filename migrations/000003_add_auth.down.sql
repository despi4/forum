-- Сначала удаляем индекс (в некоторых БД он удаляется вместе с таблицей, но явно лучше)
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_sessions_expires_at;
-- Удаляем саму таблицу
DROP TABLE IF EXISTS sessions;
