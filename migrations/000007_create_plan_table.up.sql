CREATE TABLE IF NOT EXISTS day (
    id bigserial PRIMARY KEY,
    breakfast_id bigint NOT NULL REFERENCES breakfast(id),
    am_snack_id bigint REFERENCES am_snack(id),
    lunch_id bigint NOT NULL REFERENCES lunch(id),
    pm_snack_id bigint REFERENCES pm_snack(id),
    dinner_id bigint NOT NULL REFERENCES dinner(id),
    coach bigint NOT NULL REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS plan_meal (
    id bigserial PRIMARY KEY,
    first_day bigint NOT NULL REFERENCES day(id),
    second_day bigint NOT NULL REFERENCES day(id),
    third_day bigint NOT NULL REFERENCES day(id),
    fourth_day bigint NOT NULL REFERENCES day(id),
    fifth_day bigint NOT NULL REFERENCES day(id),
    sixth_day bigint NOT NULL REFERENCES day(id),
    seventh_day bigint NOT NULL REFERENCES day(id),

    coach bigint NOT NULL  REFERENCES users(id)
);



CREATE TABLE IF NOT EXISTS user_card (
    id bigserial PRIMARY KEY,
    owner bigint NOT NULL UNIQUE REFERENCES users(id), 
    coach bigint   REFERENCES users(id), 
    current_plan  bigint REFERENCES plan_meal(id)
);
