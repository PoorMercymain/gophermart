-- +goose Up
BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS orders(num TEXT PRIMARY KEY, uploaded_at TIMESTAMP, stat TEXT, username TEXT, accrual INTEGER);
CREATE INDEX IF NOT EXISTS orders_idx ON orders USING BTREE (num, username);
CREATE TABLE IF NOT EXISTS balances(username TEXT PRIMARY KEY, balance INTEGER, withdrawn INTEGER);
CREATE INDEX IF NOT EXISTS balances_idx ON balances USING BTREE (username, balance, withdrawn);
CREATE TABLE IF NOT EXISTS withdrawals(username TEXT, order_number TEXT PRIMARY KEY, withdrawn INTEGER, processed_at TIMESTAMP);
COMMIT;

-- +goose Down
BEGIN TRANSACTION;
DROP TABLE IF EXISTS orders, balances, withdrawals;
COMMIT;