CREATE DATABASE "auth";
CREATE DATABASE "order";
CREATE DATABASE "user";

\connect auth

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS user_records (
  id BIGINT PRIMARY KEY,
  username TEXT NOT NULL,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL,
  CONSTRAINT uni_user_records_username UNIQUE (username)
);

CREATE TABLE IF NOT EXISTS service_records (
  id BIGINT PRIMARY KEY,
  secret_hash TEXT NOT NULL,
  role TEXT NOT NULL
);

INSERT INTO user_records (id, username, password_hash, role)
VALUES
  (1, 'user_all', crypt('123', gen_salt('bf')), 'user_all'),
  (2, 'user_users', crypt('123', gen_salt('bf')), 'user_users'),
  (3, 'user_orders', crypt('123', gen_salt('bf')), 'user_orders')
ON CONFLICT (id) DO NOTHING;

INSERT INTO service_records (id, secret_hash, role)
VALUES
  (1, crypt('123', gen_salt('bf')), 'api_gw'),
  (2, crypt('123', gen_salt('bf')), 'users_gw'),
  (3, crypt('123', gen_salt('bf')), 'orders_gw')
ON CONFLICT (id) DO NOTHING;
