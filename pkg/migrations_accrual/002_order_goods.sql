-- +goose Up
BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS order_goods(id SERIAL PRIMARY KEY, order_number VARCHAR(100), description VARCHAR(200), price DOUBLE PRECISION);
CREATE INDEX IF NOT EXISTS order_num_idx ON order_goods USING hash (order_number);
COMMIT;

-- +goose Down
BEGIN TRANSACTION;
DROP TABLE IF EXISTS order_goods;
COMMIT;