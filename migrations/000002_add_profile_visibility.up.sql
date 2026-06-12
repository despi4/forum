alter table users
add column visibility TEXT NOT NULL DEFAULT 'public' CHECK (visibility IN ('private', 'public'));