ALTER TABLE food ADD COLUMN IF NOT EXISTS coach bigint  REFERENCES users(id);
ALTER TABLE breakfast ADD COLUMN IF NOT EXISTS coach bigint  REFERENCES users(id);
ALTER TABLE am_snack ADD COLUMN IF NOT EXISTS coach bigint  REFERENCES users(id);
ALTER TABLE lunch ADD COLUMN IF NOT EXISTS coach bigint  REFERENCES users(id);
ALTER TABLE pm_snack ADD COLUMN IF NOT EXISTS coach bigint  REFERENCES users(id);
ALTER TABLE dinner ADD COLUMN IF NOT EXISTS coach bigint REFERENCES users(id);


