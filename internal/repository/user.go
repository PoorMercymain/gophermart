package repository

import (
	"context"
	"errors"
	"time"

	"github.com/PoorMercymain/gophermart/internal/domain"
	"github.com/PoorMercymain/gophermart/pkg/util"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO: add mutexes
type user struct {
	mongoCollection *mongo.Collection
	pgxPool         *pgxpool.Pool
}

func NewUser(mongoCollection *mongo.Collection, pgxPool *pgxpool.Pool) *user {
	return &user{mongoCollection: mongoCollection, pgxPool: pgxPool}
}

// TODO: add setting balance if registered
func (r *user) Register(ctx context.Context, user domain.User, uniqueLoginErrorChan chan error) error {
	_, err := r.mongoCollection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			uniqueLoginErrorChan <- err
			close(uniqueLoginErrorChan)
		}
	}
	return err
}

func (r *user) GetPasswordHash(ctx context.Context, login string) (string, error) {
	var user domain.User
	filter := bson.M{"login": login}
	err := r.mongoCollection.FindOne(ctx, filter).Decode(&user)
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

func (r *user) AddOrder(ctx context.Context, orderNumber int) error {
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

	_, err = tx.Exec(ctx, "INSERT INTO orders VALUES($1, $2, $3, $4)", orderNumber, time.Now(), "NEW", ctx.Value(domain.Key("login")))
	if err != nil {
		var login string
		util.LogInfoln(pgErr)
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			tx.Rollback(ctx)
			util.LogInfoln(err)
			conn.QueryRow(ctx, "SELECT username FROM orders WHERE num = $1", orderNumber).Scan(&login)
			util.LogInfoln(login)
			if login == ctx.Value(domain.Key("login")).(string) {
				util.LogInfoln(domain.ErrorAlreadyRegistered.Error())
				return domain.ErrorAlreadyRegistered
			}
			util.LogInfoln(domain.ErrorAlreadyRegisteredByAnotherUser.Error())
			return domain.ErrorAlreadyRegisteredByAnotherUser
		}
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

	//TODO: add limit
	rows, err := conn.Query(ctx, "SELECT num, stat, uploaded_at FROM orders WHERE username = $1 ORDER BY uploaded_at DESC", ctx.Value(domain.Key("login")))
	if err != nil {
		util.LogInfoln(err)
		return nil, err
	}

	var order domain.Order
	orders := make([]domain.Order, 0)

	for rows.Next() {
		rows.Scan(&order.Number, &order.Status, &order.UploadedAt) // TODO: need to set actual accrual here
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
