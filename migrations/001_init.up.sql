CREATE TABLE team (
    id   SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE "user" (
    id        TEXT UNIQUE PRIMARY KEY,
    name      TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    team_id   INT NOT NULL REFERENCES team(id)
);

CREATE TABLE pull_request (
    id         TEXT PRIMARY KEY,
    title      TEXT NOT NULL,
    author_id  TEXT NOT NULL REFERENCES "user"(id),
    status     TEXT NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NULL
);

CREATE TABLE pull_request_reviewer (
    pr_id   TEXT NOT NULL REFERENCES pull_request(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES "user" (id),
    PRIMARY KEY (pr_id, user_id)
);
