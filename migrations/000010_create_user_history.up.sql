CREATE TABLE IF NOT EXISTS user_history(
    id bigserial PRIMARY KEY,
    plan_done  bigint  NOT NULL REFERENCES plan_meal(id),
    from_date timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    to_date timestamp(0) with time zone,
    weight_start int NOT NULL DEFAULT 0,
    weight_finish int NOT NULL DEFAULT 0,
    is_now boolean NOT NULL DEFAULT true,
    owner bigint NOT NULL REFERENCES users(id)
);

ALTER TABLE user_card ADD COLUMN IF NOT EXISTS current_weight int NOT NULL DEFAULT 0;