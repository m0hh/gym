CREATE EXTENSION IF NOT EXISTS citext;
CREATE TYPE role_enum  AS ENUM ('admin','coach','trainee','gym');
CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    activated bool NOT NULL,
    version integer NOT NULL DEFAULT 1,
    role role_enum NOT NULL
);