ALTER TABLE exercise ADD COLUMN IF NOT EXISTS coach bigint NOT NULL REFERENCES users(id);
ALTER TABLE exercise_day ADD COLUMN IF NOT EXISTS coach bigint NOT NULL REFERENCES users(id);
ALTER TABLE exercise_plan ADD COLUMN IF NOT EXISTS coach bigint NOT NULL REFERENCES users(id);