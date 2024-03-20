
CREATE TABLE IF NOT EXISTS exercise_name(
    id bigserial PRIMARY KEY,
    name text NOT NULL
);

CREATE TABLE IF NOT EXISTS exercise(
    id bigserial PRIMARY KEY,
    name bigint  NOT NULL REFERENCES exercise_name(id),
    sets int NOT NULL,
    reps int NOT NULL
);


CREATE TABLE IF NOT EXISTS exercise_day(
    id bigserial PRIMARY KEY,
    name text NOT NULL
);

CREATE TABLE IF NOT EXISTS exercises_to_day (
    exercise_id bigint NOT NULL REFERENCES exercise(id),
    day_id bigint NOT NULL REFERENCES exercise_day(id),
    PRIMARY KEY (exercise_id, day_id)
);

CREATE TABLE IF NOT EXISTS exercise_plan(
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    how_to text NOT NULL
);

CREATE TABLE IF NOT EXISTS days_to_plan (
    plan_id bigint NOT NULL REFERENCES exercise_plan(id),
    day_id bigint NOT NULL REFERENCES exercise_day(id),
    PRIMARY KEY (plan_id, day_id)
);