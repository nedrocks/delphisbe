ALTER TABLE posts ADD COLUMN IF NOT EXISTS post_type varchar(20) DEFAULT 'STANDARD' not null;