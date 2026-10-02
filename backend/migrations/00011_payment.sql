-- +goose Up
SET LOCAL lock_timeout = '5s';

ALTER TABLE courses
    ADD COLUMN price_minor bigint NULL CHECK (price_minor IS NULL OR price_minor > 0),
    ADD COLUMN currency text NULL CHECK (currency IS NULL OR currency = 'UAH'),
    ADD CONSTRAINT courses_price_currency_check
        CHECK ((price_minor IS NULL) = (currency IS NULL));

ALTER TABLE course_lessons
    ADD COLUMN is_preview boolean NOT NULL DEFAULT false;

CREATE TABLE orders (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id),
    course_id uuid NOT NULL REFERENCES courses(id),
    amount_minor bigint NOT NULL CHECK (amount_minor > 0),
    currency text NOT NULL CHECK (currency = 'UAH'),
    status text NOT NULL CHECK (status IN ('pending', 'paid', 'failed')),
    idempotency_key text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    paid_at timestamptz NULL,
    provider text NULL,
    provider_ref text NULL,
    UNIQUE (user_id, idempotency_key)
);

CREATE TABLE course_grants (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id),
    course_id uuid NOT NULL REFERENCES courses(id),
    source text NOT NULL CHECK (source IN ('purchase', 'admin', 'grandfather')),
    order_id uuid NULL REFERENCES orders(id),
    state text NOT NULL CHECK (state IN ('active', 'revoked')),
    granted_at timestamptz NOT NULL DEFAULT now(),
    granted_by uuid NULL REFERENCES users(id),
    CHECK ((source = 'purchase') = (order_id IS NOT NULL))
);

CREATE UNIQUE INDEX course_grants_one_active_per_user_course
    ON course_grants (user_id, course_id)
    WHERE state = 'active';

CREATE TABLE payment_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    provider text NOT NULL,
    event_id text NOT NULL,
    order_id uuid NULL REFERENCES orders(id),
    type text NOT NULL,
    amount_minor bigint NULL,
    received_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (provider, event_id)
);

CREATE TABLE grant_audit (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    grant_id uuid NOT NULL REFERENCES course_grants(id),
    actor_id uuid NULL,
    action text NOT NULL,
    reason text NULL,
    payment_event_id uuid NULL REFERENCES payment_events(id),
    created_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE grant_audit;
DROP TABLE payment_events;
DROP TABLE course_grants;
DROP TABLE orders;

ALTER TABLE course_lessons
    DROP COLUMN is_preview;

ALTER TABLE courses
    DROP CONSTRAINT courses_price_currency_check,
    DROP COLUMN currency,
    DROP COLUMN price_minor;
