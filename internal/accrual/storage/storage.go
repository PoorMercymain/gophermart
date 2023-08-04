package storage

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"

	"github.com/PoorMercymain/gophermart/internal/accrual/domain"
	"github.com/PoorMercymain/gophermart/internal/accrual/interfaces"
	"github.com/PoorMercymain/gophermart/pkg/util"
)

type dbStorage struct {
	pgxPool *pgxpool.Pool
	mutex   *sync.Mutex
}

func NewDBStorage(DSN string) (storage interfaces.Storage, err error) {
	pg, err := sql.Open("pgx/v5", DSN)
	if err != nil {
		util.GetLogger().Infoln(err)
		return
	}
	err = goose.SetDialect("postgres")
	if err != nil {
		util.GetLogger().Infoln(err)
		return
	}

	err = pg.PingContext(context.Background())
	if err != nil {
		util.GetLogger().Infoln(err)

		for _, retryInterval := range domain.RepeatedAttemptsIntervals {
			time.Sleep(*retryInterval)
			err = pg.PingContext(context.Background())
			if err != nil {
				util.GetLogger().Infoln(err)
			} else {
				util.GetLogger().Infoln("ping succesful")
				break
			}

		}
		if err != nil {
			return
		}
	}

	err = goose.Run("up", pg, "./pkg/migrations_accrual")
	if err != nil {
		util.GetLogger().Infoln(err)
		return
	}
	pg.Close()

	config, err := pgxpool.ParseConfig(DSN)
	if err != nil {
		util.GetLogger().Infoln("Error parsing DSN:", err)
		return
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		util.GetLogger().Infoln("Error creating pgxpool:", err)
		return
	}

	util.GetLogger().Infoln("Pool created", pool, err)

	storage = &dbStorage{pgxPool: pool, mutex: &sync.Mutex{}}
	return
}

func (dbs *dbStorage) StoreGoodsReward(ctx context.Context, goods *domain.Goods) (err error) {

	conn, err := dbs.pgxPool.Acquire(ctx)
	if err != nil {
		return
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback(ctx)

	var pgErr *pgconn.PgError

	_, err = tx.Exec(ctx, "INSERT INTO goods (id, match, reward, reward_type) "+
		"VALUES(DEFAULT, $1, $2, $3)", goods.Match, goods.Reward, goods.RewardType)

	if err != nil {

		util.GetLogger().Infoln(err)
		util.GetLogger().Infoln(pgErr)
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			tx.Rollback(ctx)
			util.GetLogger().Infoln(err)
			return domain.ErrorMatchAlreadyRegistered
		}
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		util.GetLogger().Infoln(err)
		util.GetLogger().Infoln(pgErr)
	}
	return
}
func (dbs *dbStorage) StoreOrder(ctx context.Context, order *domain.OrderRecord) (err error) {

	conn, err := dbs.pgxPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var pgErr *pgconn.PgError

	_, err = tx.Exec(ctx, "INSERT INTO orders (id, number, status, accrual) VALUES"+
		"(DEFAULT, $1, $2, $3)", order.Number, order.Status, order.Accrual)

	if err != nil {
		util.GetLogger().Infoln(pgErr)
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			util.GetLogger().Infoln(err)
			return domain.ErrorOrderAlreadyProcessing
		}
		return err
	}

	return tx.Commit(ctx)
}

func (dbs *dbStorage) UpdateOrder(ctx context.Context, order *domain.OrderRecord) (err error) {

	conn, err := dbs.pgxPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var pgErr *pgconn.PgError

	_, err = tx.Exec(ctx, "UPDATE orders SET status = $1, accrual = $2 "+
		"WHERE number = $3", order.Status, order.Accrual, order.Number)

	if err != nil {
		util.GetLogger().Infoln(pgErr)
		return err
	}

	return tx.Commit(ctx)
}

func (dbs *dbStorage) GetOrder(ctx context.Context, number *string) (order *domain.OrderRecord, err error) {
	conn, err := dbs.pgxPool.Acquire(ctx)
	if err != nil {
		return
	}
	defer conn.Release()

	order = &domain.OrderRecord{}

	row := conn.QueryRow(ctx, "SELECT number, status, accrual FROM orders WHERE number = $1", *number)

	err = row.Scan(&order.Number, &order.Status, &order.Accrual)

	return
}

//TODO: cache it, update cache on StoreGoodsReward

func (dbs *dbStorage) GetGoods(ctx context.Context) (goods []*domain.Goods, err error) {
	conn, err := dbs.pgxPool.Acquire(ctx)
	if err != nil {
		return
	}
	defer conn.Release()

	var totalRows int

	err = conn.QueryRow(ctx, "SELECT count(*) FROM goods").
		Scan(&totalRows)

	if totalRows == 0 {
		util.GetLogger().Infoln("no goods registered")
		return
	}

	rows, err := conn.Query(ctx, "SELECT match, reward, reward_type FROM goods")
	if err != nil {
		util.GetLogger().Infoln(err)
		return
	}

	goods = make([]*domain.Goods, totalRows)
	counter := 0
	for rows.Next() {
		var goodsRecord = &domain.Goods{}
		err = rows.Scan(&goodsRecord.Match, &goodsRecord.Reward, &goodsRecord.RewardType)
		if err != nil {
			util.GetLogger().Infoln(err)
			return
		}
		goods[counter] = goodsRecord
		counter++
	}

	return
}

func (dbs *dbStorage) StoreOrderGoods(ctx context.Context, order *domain.Order) (err error) {

	conn, err := dbs.pgxPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	batch := &pgx.Batch{}

	for _, value := range order.Goods {
		batch.Queue("INSERT INTO order_goods (id, order_number, description, price) VALUES(DEFAULT, $1, $2, $3)",
			order.Number, value.Description, value.Price)
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		util.GetLogger().Infoln(err)
		return err
	}
	defer tx.Rollback(ctx)

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for range order.Goods {
		_, err = br.Exec()
		if err != nil {
			util.GetLogger().Infoln(err)
			return
		}
	}

	err = br.Close()
	if err != nil {
		util.GetLogger().Infoln(err)
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		util.GetLogger().Infoln(err)
		return err
	}

	return
}
func (dbs *dbStorage) GetOrderGoods(ctx context.Context, orderNumber *string) (orderGoods []*domain.OrderGoods, err error) {
	conn, err := dbs.pgxPool.Acquire(ctx)
	if err != nil {
		return
	}
	defer conn.Release()

	var totalRows int

	err = conn.QueryRow(ctx, "SELECT count(*) FROM order_goods WHERE order_number = $1", *orderNumber).
		Scan(&totalRows)

	if totalRows == 0 {
		util.GetLogger().Infoln("no goods for order", *orderNumber)
		return
	}

	rows, err := conn.Query(ctx, "SELECT description, price FROM order_goods WHERE order_number = $1", *orderNumber)
	if err != nil {
		util.GetLogger().Infoln(err)
		return
	}

	orderGoods = make([]*domain.OrderGoods, totalRows)
	counter := 0
	for rows.Next() {
		var goodsRecord = &domain.OrderGoods{}
		err = rows.Scan(&goodsRecord.Description, &goodsRecord.Price)
		if err != nil {
			util.GetLogger().Infoln(err)
			return
		}
		orderGoods[counter] = goodsRecord
		counter++
	}

	return
}
func (dbs *dbStorage) GetUnprocessedOrders(ctx context.Context) (orders []*domain.OrderRecord, err error) {
	conn, err := dbs.pgxPool.Acquire(ctx)
	if err != nil {
		return
	}
	defer conn.Release()

	var totalRows int

	err = conn.QueryRow(ctx, "SELECT count(*) FROM orders WHERE status = $1 OR status = $2", domain.OrderStatusProcessing, domain.OrderStatusRegistered).
		Scan(&totalRows)

	if totalRows == 0 {
		util.GetLogger().Infoln("no unprocessed orders found")
		return
	}

	rows, err := conn.Query(ctx, "SELECT number FROM orders WHERE status = $1 OR status = $2", domain.OrderStatusProcessing, domain.OrderStatusRegistered)
	if err != nil {
		util.GetLogger().Infoln(err)
		return
	}

	orders = make([]*domain.OrderRecord, totalRows)
	counter := 0
	for rows.Next() {
		var orderRecord = &domain.OrderRecord{}
		err = rows.Scan(&orderRecord.Number)
		if err != nil {
			util.GetLogger().Infoln(err)
			return
		}
		orders[counter] = orderRecord
		counter++
	}

	return
}

func (dbs *dbStorage) ClosePool() (err error) {
	dbs.pgxPool.Close()
	return
}

func (dbs *dbStorage) Ping(ctx context.Context) (err error) {
	conn, err := dbs.pgxPool.Acquire(ctx)
	if err != nil {
		return
	}
	defer conn.Release()

	err = conn.Ping(ctx)
	if err != nil {
		return
	}
	return
}
