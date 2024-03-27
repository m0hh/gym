
ALTER TABLE user_card ADD COLUMN IF NOT EXISTS current_exercise_plan bigint REFERENCES exercise_plan(id);

ALTER TABLE user_history ADD COLUMN IF NOT EXISTS exercise_plan_done bigint REFERENCES exercise_plan(id);