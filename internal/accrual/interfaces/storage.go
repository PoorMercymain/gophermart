package interfaces

import (
	"context"

	"github.com/PoorMercymain/gophermart/internal/accrual/domain"
)

type Storage interface {
	StoreGoodsReward(ctx context.Context, goods *domain.Goods) error
	StoreOrder(ctx context.Context, order *domain.OrderRecord) error
	UpdateOrder(ctx context.Context, order *domain.OrderRecord) error
	GetOrder(ctx context.Context, num *string) (*domain.OrderRecord, error)
	GetGoods(ctx context.Context) ([]*domain.Goods, error)

	StoreOrderGoods(ctx context.Context, order *domain.Order) error
	GetOrderGoods(ctx context.Context, num *string) ([]*domain.OrderGoods, error)
	GetUnprocessedOrders(ctx context.Context) ([]*domain.OrderRecord, error)

	Ping(ctx context.Context) error
	ClosePool() error
}
