-- Своя база на сервис: у goose общая таблица миграций, в одной базе сервисы бы её делили.
-- Идемпотентно: можно запускать повторно (в docker-entrypoint-initdb.d выполняется только на пустом томе).
SELECT format('CREATE DATABASE %I', d)
FROM (VALUES ('auth'), ('users'), ('boards')) AS t(d)
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = d)
          \gexec
