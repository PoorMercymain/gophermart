package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/jackc/pgerrcode"
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
		fmt.Println(err)
		return
	}
	err = goose.SetDialect("postgres")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = pg.PingContext(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	err = goose.Run("up", pg, "./pkg/migrations_accrual")
	if err != nil {
		fmt.Println(err)
		return
	}
	pg.Close()

	config, err := pgxpool.ParseConfig(DSN)
	if err != nil {
		fmt.Println("Error parsing DSN:", err)
		return
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		fmt.Println("Error creating pgxpool:", err)
		return
	}

	fmt.Println("Pool created", pool, err)

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

	currInsertValues := fmt.Sprintf("INSERT INTO goods (id, match, reward, reward_type) "+
		"VALUES(DEFAULT, '%v', %v, '%v')", goods.Match, goods.Reward, goods.RewardType)

	_, err = tx.Exec(ctx, currInsertValues)

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

	currInsertValues := fmt.Sprintf("INSERT INTO orders (id, number, status, accrual) VALUES"+
		"(DEFAULT, '%v', '%v', %v)", order.Number, order.Status, order.Accrual)

	_, err = tx.Exec(ctx, currInsertValues)

	if err != nil {
		util.GetLogger().Infoln(pgErr)
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			tx.Rollback(ctx)
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

	currInsertValues := fmt.Sprintf("UPDATE orders SET status = '%v', accrual = %v "+
		"WHERE number = '%v'", order.Status, order.Accrual, order.Number)

	_, err = tx.Exec(ctx, currInsertValues)

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

	var goodsRecord = &domain.Goods{}
	goods = make([]*domain.Goods, totalRows)
	counter := 0

	for rows.Next() {
		rows.Scan(goodsRecord.Match, goodsRecord.Reward, goodsRecord.RewardType)
		goods[counter] = goodsRecord
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
