CREATE TABLE public.private_data (
      id UUID PRIMARY KEY NOT NULL,
      user_id UUID NOT NULL,
      data BYTEA NOT NULL,
      created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
      deleted_at TIMESTAMPTZ DEFAULT NULL
);