package interfaces

import "github.com/PoorMercymain/gophermart/internal/accrual/domain"
import "context"

type Storage interface {
	StoreGoodsReward(ctx context.Context, goods *domain.Goods) error
	StoreOrder(ctx context.Context, order *domain.OrderRecord) error
	UpdateOrder(ctx context.Context, order *domain.OrderRecord) error
	GetOrder(ctx context.Context, num *string) (*domain.OrderRecord, error)
	GetGoods(ctx context.Context) ([]*domain.Goods, error)

	Ping(ctx context.Context) error
	ClosePool() error
}
