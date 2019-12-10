CREATE USER accesswriter;
CREATE DATABASE access_postgres_db owner accesswriter;
GRANT ALL PRIVILEGES ON DATABASE access_postgres_db TO accesswriter;
\connect access_postgres_db accesswriter
CREATE TABLE IF NOT EXISTS access
(
	id SERIAL PRIMARY KEY,
	ip VARCHAR(45),
	access_time TIMESTAMPTZ
);
ALTER USER accesswriter WITH PASSWORD 'hu8jmn3';
