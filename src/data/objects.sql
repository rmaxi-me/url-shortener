CREATE TABLE urls
(
    id          BIGSERIAL PRIMARY KEY,
    source      TEXT        NOT NULL,
    target      TEXT        NOT NULL,
    delete_key  TEXT        NOT NULL,
    serve_count INT                  DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
)