ALTER TABLE participants ADD COLUMN IF NOT EXISTS is_banned boolean DEFAULT False not null;