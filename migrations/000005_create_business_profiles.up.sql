-- +migrate Up
CREATE TABLE IF NOT EXISTS business_profiles (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name text NOT NULL,
    logo_url text NOT NULL,
    rccm text,
    address text,
    phone text,
    email text,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);




