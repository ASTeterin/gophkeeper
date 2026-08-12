CREATE TABLE public.private_data (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     user_id UUID NOT NULL,
     data_key VARCHAR(255) NOT NULL,
     description VARCHAR(255),
     data BYTEA NOT NULL,
     CONSTRAINT user_id_data_key UNIQUE (user_id, data_key)
);