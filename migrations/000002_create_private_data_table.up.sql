CREATE TABLE public.private_data (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    data_key VARCHAR(255) NOT NULL UNIQUE,
    description VARCHAR(255),
    data BYTEA NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ DEFAULT NULL
);