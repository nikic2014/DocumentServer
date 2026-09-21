CREATE TABLE IF NOT EXISTS public.users (
    id       SERIAL PRIMARY KEY,
    login    TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS documents (
     id           SERIAL PRIMARY KEY,
     owner_login  TEXT NOT NULL REFERENCES users(login) ON DELETE CASCADE,
     name         TEXT NOT NULL,
     mime         TEXT,
     file_path    TEXT,
     is_public    BOOLEAN NOT NULL DEFAULT false,
     grant_logins JSONB NOT NULL DEFAULT '[]',
     json_data    JSONB,
     created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
