CREATE TABLE dedupe.file_entry (
    id uuid PRIMARY KEY default gen_random_uuid(),
    scan_id uuid not null,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL default now(),
    name character varying not null,
    size BIGINT not null,
    mode BIGINT not null,
    mod_time TIMESTAMP WITH TIME ZONE not null,
    is_dir boolean not null,
    hash varchar(64) not null,
    error character varying,
);