

CREATE TABLE IF NOT EXISTS food (
    id bigserial PRIMARY KEY,
    food_name text NOT NULL,
    serving text NOT NULL
);

CREATE TABLE IF NOT EXISTS breakfast (
    id bigserial PRIMARY KEY,
    calories integer
);

CREATE TABLE IF NOT EXISTS breakfast_food (
    breakfast_id bigint REFERENCES breakfast(id),
    food_id bigint REFERENCES food(id),
    PRIMARY KEY (breakfast_id, food_id)
);

CREATE TABLE IF NOT EXISTS am_snack (
    id bigserial PRIMARY KEY,
    calories integer
);

CREATE TABLE IF NOT EXISTS am_snack_food (
    am_snack_id bigint REFERENCES am_snack(id),
    food_id bigint REFERENCES food(id),
    PRIMARY KEY (am_snack_id, food_id)
);

CREATE TABLE IF NOT EXISTS lunch (
    id bigserial PRIMARY KEY,
    calories integer

);

CREATE TABLE IF NOT EXISTS lunch_food (
    lunch_id bigint REFERENCES lunch(id),
    food_id bigint REFERENCES food(id),
    PRIMARY KEY (lunch_id, food_id)
);


CREATE TABLE IF NOT EXISTS pm_snack (
    id bigserial PRIMARY KEY,
    calories integer

);

CREATE TABLE IF NOT EXISTS pm_snack_food (
    pm_snack_id bigint REFERENCES pm_snack(id),
    food_id bigint REFERENCES food(id),
    PRIMARY KEY (pm_snack_id, food_id)
);


CREATE TABLE IF NOT EXISTS dinner (
    id bigserial PRIMARY KEY,
    calories integer

);

CREATE TABLE IF NOT EXISTS dinner_food (
    dinner_id bigint REFERENCES dinner(id),
    food_id bigint REFERENCES food(id),
    PRIMARY KEY (dinner_id, food_id)
);