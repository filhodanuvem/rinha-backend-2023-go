ALTER SYSTEM SET max_connections = 300;

-- Use an extension to enable trigram similarity search and improve LIKE performance
-- https://www.postgresql.org/docs/current/runtime-config-connection.htmlhttps://mazeez.dev/posts/pg-trgm-similarity-search-and-fast-like
CREATE EXTENSION pg_trgm;

ALTER DATABASE rinha SET synchronous_commit=OFF;
-- using 25% of memory as suggested in the docs:
--    https://www.postgresql.org/docs/9.1/runtime-config-resource.html
ALTER SYSTEM SET shared_buffers TO "425MB";

-- debug slow queries, run \d pg_stat_statements
-- docs: 
--    https://www.postgresql.org/docs/current/pgstatstatements.html
-- CREATE EXTENSION pg_stat_statements;
-- ALTER SYSTEM SET shared_preload_libraries = 'pg_stat_statements';


create table if not exists pessoas(
  id uuid not null primary key,
  apelido varchar(64) not null unique,
  nome varchar(200) not null,
  nascimento timestamp not null,
  stack VARCHAR[] null default '{}',
  search_index varchar(1200) not null,
  created_at timestamp not null default current_timestamp
);

CREATE INDEX pessoas_search_index_idx ON pessoas USING gin (search_index gin_trgm_ops);