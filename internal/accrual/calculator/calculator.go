package calculator

import (
	"context"
	"strings"

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

	goodsRewards, err := storage.GetGoods(ctx)
	if err != nil {
		util.GetLogger().Infoln(err)
		cancelCtx(err)
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
					orderRecord.Accrual += currReward.Reward * currGoods.Price
				}
				break
			}
		}

	}
	orderRecord.Status = domain.OrderStatusProcessed

	err = storage.UpdateOrder(ctx, &orderRecord)
	if err != nil {
		util.GetLogger().Infoln(err)
		cancelCtx(err)
		return
	}

	util.GetLogger().Infoln(orderRecord)
}
