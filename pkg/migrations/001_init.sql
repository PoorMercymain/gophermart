-- +goose Up
BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS orders(num INTEGER PRIMARY KEY, uploaded_at TIMESTAMP, stat TEXT, username TEXT);
CREATE TABLE IF NOT EXISTS balances(username TEXT PRIMARY KEY, balance INTEGER, withdrawn INTEGER);
COMMIT;

-- +goose Down
BEGIN TRANSACTION;
DROP TABLE IF EXISTS orders, balances;
COMMIT;