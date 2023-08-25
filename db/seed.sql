CREATE EXTENSION pg_trgm;

create table if not exists pessoas(
  id uuid not null primary key,
  apelido varchar(64) not null unique,
  nome varchar(200) not null,
  nascimento timestamp not null,
  stack VARCHAR[] null default '{}',
  created_at timestamp not null default current_timestamp,
  search_index varchar(1200) not null
);

CREATE INDEX pessoas_search_index_idx ON pessoas USING gin (search_index gin_trgm_ops);