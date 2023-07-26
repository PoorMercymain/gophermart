package calculator

import (
	"context"
	"math/rand"
	"time"

	"github.com/PoorMercymain/gophermart/internal/accrual/domain"
	"github.com/PoorMercymain/gophermart/internal/accrual/interfaces"
	"github.com/PoorMercymain/gophermart/pkg/util"
)

func CalculateAccrual(ctx context.Context, order *domain.Order, storage interfaces.Storage) {

	_, cancelCtx := context.WithCancelCause(ctx)

	var orderRecord = domain.OrderRecord{
		Number: order.Number,
		Status: domain.OrderStatusProcessing,
	}

	util.GetLogger().Infoln(orderRecord)

	err := storage.UpdateOrder(ctx, &orderRecord)
	if err != nil {
		util.GetLogger().Infoln(err)
		cancelCtx(err)
		return
	}

	//TODO: calculate real accrual, it's a stub
	const calculateTime = 15
	time.Sleep(time.Duration(calculateTime) * time.Second)

	orderRecord.Status = domain.OrderStatusProcessed
	orderRecord.Accrual = 1000 * rand.Float64()

	err = storage.UpdateOrder(ctx, &orderRecord)
	if err != nil {
		util.GetLogger().Infoln(err)
		cancelCtx(err)
		return
	}

	util.GetLogger().Infoln(orderRecord)
}
