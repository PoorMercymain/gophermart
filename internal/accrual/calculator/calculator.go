package calculator

import (
	"context"
	"strings"
	"sync"

	"github.com/PoorMercymain/gophermart/internal/accrual/domain"
	"github.com/PoorMercymain/gophermart/internal/accrual/interfaces"
	"github.com/PoorMercymain/gophermart/pkg/util"
)

func ProcessUnprocessedOrders(ctx context.Context, storage interfaces.Storage, wg *sync.WaitGroup) (err error) {

	orders, err := storage.GetUnprocessedOrders(ctx)
	if err != nil {
		util.GetLogger().Infoln(err)
		return
	}

	for _, order := range orders {
		util.GetLogger().Infoln("processing unprocessed order ", order.Number)
		wg.Add(1)
		go processOrder(ctx, &order.Number, storage, wg)
	}

	return
}

func processOrder(ctx context.Context, orderNumber *string, storage interfaces.Storage, wg *sync.WaitGroup) (err error) {

	defer wg.Done()
	orderGoods, err := storage.GetOrderGoods(ctx, orderNumber)
	if err != nil {
		util.GetLogger().Infoln(err)
		return
	}

	order := &domain.Order{
		Number: *orderNumber,
		Goods:  orderGoods,
	}
	wg.Add(1)
	err = CalculateAccrual(ctx, order, storage, wg)
	if err != nil {
		util.GetLogger().Infoln(err)
		return
	}

	return
}

func CalculateAccrual(ctx context.Context, order *domain.Order, storage interfaces.Storage, wg *sync.WaitGroup) (err error) {

	defer wg.Done()

	var orderRecord = domain.OrderRecord{
		Number: order.Number,
		Status: domain.OrderStatusProcessing,
	}

	util.GetLogger().Infoln("calculating accrual for order ", orderRecord)

	err = storage.UpdateOrder(ctx, &orderRecord)
	if err != nil {
		util.GetLogger().Infoln(err)
		return
	}

	goodsRewards, err := storage.GetGoods(ctx)
	if err != nil {
		util.GetLogger().Infoln(err)
		return
	}

	for _, currGoods := range order.Goods {
		//find first matching reward for current goods
		for _, currReward := range goodsRewards {
			if strings.Contains(currGoods.Description, currReward.Match) {
				switch currReward.RewardType {
				case domain.RewardTypePt:
					orderRecord.Accrual += currReward.Reward
				case domain.RewardTypePercent:
					orderRecord.Accrual += currReward.Reward * currGoods.Price / 100
				}
				break
			}
		}

	}
	orderRecord.Status = domain.OrderStatusProcessed
	err = storage.UpdateOrder(ctx, &orderRecord)
	if err != nil {
		util.GetLogger().Infoln(err)
		return
	}

	util.GetLogger().Infoln("calculated accrual for order ", orderRecord)
	return
}
