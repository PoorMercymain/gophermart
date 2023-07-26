package repository

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/PoorMercymain/gophermart/internal/domain"
	"github.com/PoorMercymain/gophermart/pkg/util"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type user struct {
	mongo   *mongoWithMutex
	pgxPool *pgxpool.Pool
}

type mongoWithMutex struct {
	mongoCollection *mongo.Collection
	*sync.Mutex
}

func NewUser(mongoCollection *mongo.Collection, pgxPool *pgxpool.Pool) *user {
	return &user{mongo: &mongoWithMutex{mongoCollection, &sync.Mutex{}}, pgxPool: pgxPool}
}

func (r *user) Register(ctx context.Context, user domain.User, uniqueLoginErrorChan chan error) error {
	r.mongo.Lock()
	_, err := r.mongo.mongoCollection.InsertOne(ctx, user)
	r.mongo.Unlock()
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			uniqueLoginErrorChan <- err
			close(uniqueLoginErrorChan)
		}
		return err
	}

	conn, err := r.pgxPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, "INSERT INTO balances VALUES($1, $2, $3)", user.Login, 0, 0)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *user) GetPasswordHash(ctx context.Context, login string) (string, error) {
	var user domain.User
	filter := bson.M{"login": login}
	r.mongo.Lock()
	err := r.mongo.mongoCollection.FindOne(ctx, filter).Decode(&user)
	r.mongo.Unlock()
	if err != nil {
		return "", err
	}
	return user.Password, nil
}

func (r *user) Ping(ctx context.Context) error {
	conn, err := r.pgxPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	err = conn.Ping(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *user) AddOrder(ctx context.Context, orderNumber string) error {
	conn, err := r.pgxPool.Acquire(ctx)
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

	_, err = tx.Exec(ctx, "INSERT INTO orders VALUES($1, $2, $3, $4, $5)", orderNumber, time.Now(), "NEW", ctx.Value(domain.Key("login")), 0)
	if err != nil {
		var login string
		util.GetLogger().Infoln(pgErr)
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			tx.Rollback(ctx)
			util.GetLogger().Infoln(err)
			conn.QueryRow(ctx, "SELECT username FROM orders WHERE num = $1", orderNumber).Scan(&login)
			util.GetLogger().Infoln(login)
			if login == ctx.Value(domain.Key("login")).(string) {
				util.GetLogger().Infoln(domain.ErrorAlreadyRegistered.Error())
				return domain.ErrorAlreadyRegistered
			}
			util.GetLogger().Infoln(domain.ErrorAlreadyRegisteredByAnotherUser.Error())
			return domain.ErrorAlreadyRegisteredByAnotherUser
		}
		util.GetLogger().Infoln(err)
		return err
	}

	return tx.Commit(ctx)
}

func (r *user) ReadOrders(ctx context.Context) ([]domain.Order, error) {
	conn, err := r.pgxPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, "SELECT num, stat, uploaded_at, accrual FROM orders WHERE username = $1 ORDER BY uploaded_at DESC LIMIT 15 OFFSET $2", ctx.Value(domain.Key("login")), ((ctx.Value(domain.Key("page")).(int))-1)*15)
	if err != nil {
		util.GetLogger().Infoln(err)
		return nil, err
	}

	var order domain.Order
	orders := make([]domain.Order, 0)
	var accrual int

	for rows.Next() {
		rows.Scan(&order.Number, &order.Status, &order.UploadedAt, &accrual)
		util.GetLogger().Infoln("пам-пам", order, accrual)
		if accrual != 0 {
			order.Accrual.Money = accrual
		}
		order.UploadedAtString = order.UploadedAt.Format(time.RFC3339)
		orders = append(orders, order)
	}

	return orders, nil
}

func (r *user) ReadBalance(ctx context.Context) (domain.Balance, error) {
	conn, err := r.pgxPool.Acquire(ctx)
	if err != nil {
		return domain.Balance{}, err
	}
	defer conn.Release()

	var balance domain.Balance

	conn.QueryRow(ctx, "SELECT balance, withdrawn FROM balances WHERE username = $1", ctx.Value(domain.Key("login"))).Scan(&balance.Balance, &balance.Withdrawn)

	return balance, nil
}

func (r *user) AddWithdrawal(ctx context.Context, withdrawal domain.Withdrawal) error {
	conn, err := r.pgxPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var balance domain.Balance

	tx.QueryRow(ctx, "SELECT balance, withdrawn FROM balances WHERE username = $1", ctx.Value(domain.Key("login"))).Scan(&balance.Balance, &balance.Withdrawn)

	if balance.Balance < withdrawal.WithdrawalAmount.Withdrawal {
		return domain.ErrorNotEnoughPoints
	}

	_, err = tx.Exec(ctx, "UPDATE balances SET balance = balance - $1, withdrawn = withdrawn + $2 WHERE username = $3", withdrawal.WithdrawalAmount, withdrawal.WithdrawalAmount, ctx.Value(domain.Key("login")))
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "INSERT INTO withdrawals VALUES($1, $2, $3, $4)", ctx.Value(domain.Key("login")), withdrawal.OrderNumber, withdrawal.WithdrawalAmount, time.Now())
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *user) UpdateOrder(ctx context.Context, order domain.AccrualOrder) error {
	conn, err := r.pgxPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var accrual int
	if order.Status == "PROCESSED" {
		conn.QueryRow(ctx, "SELECT accrual FROM orders WHERE num = $1", order.Order).Scan(&accrual)
		accrualDiff := order.Accrual.Accrual - accrual
		if accrualDiff > 0 {
			_, err = tx.Exec(ctx, "UPDATE balances SET balance = balance + $1 WHERE username = $2", accrualDiff, ctx.Value(domain.Key("login")))
			if err != nil {
				util.GetLogger().Infoln(err)
				return err
			}
		}
	}

	_, err = tx.Exec(ctx, "UPDATE orders SET stat = $1, accrual = $2 WHERE num = $3", order.Status, order.Accrual.Accrual, order.Order)
	if err != nil {
		return err
	}

	util.GetLogger().Infoln("it works,", order, "is now updated!")
	return tx.Commit(ctx)
}
