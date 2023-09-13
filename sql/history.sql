CREATE TABLE history (
    id          SERIAL PRIMARY KEY,
    ts          TIMESTAMPTZ NOT NULL DEFAULT now(),
    username    TEXT,
    prompt      TEXT,
    result      JSONB
);
