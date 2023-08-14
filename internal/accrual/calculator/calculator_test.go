package calculator

import (
	"context"
	"github.com/PoorMercymain/gophermart/internal/accrual/domain"
	"github.com/PoorMercymain/gophermart/internal/accrual/interfaces/mocks"
	"github.com/PoorMercymain/gophermart/pkg/util"
	"github.com/golang/mock/gomock"
	"sync"
	"testing"
)

func TestCalculateAccrual(t *testing.T) {
	ctx := context.TODO()
	wg := &sync.WaitGroup{}
	util.InitLogger()
	mockRepo := newMockupRepo(t)

	type args struct {
		order *domain.Order
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test calculator",
			args: args{
				order: &domain.Order{
					Number: "11112",
					Goods: []*domain.OrderGoods{
						{Description: "test", Price: 10},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		wg.Add(1)
		t.Run(tt.name, func(t *testing.T) {
			if err := CalculateAccrual(ctx, tt.args.order, mockRepo, wg); (err != nil) != tt.wantErr {
				t.Errorf("CalculateAccrual() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func TestProcessUnprocessedOrders(t *testing.T) {
	ctx := context.TODO()
	wg := &sync.WaitGroup{}
	util.InitLogger()
	mockRepo := newMockupRepo(t)

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Test unprocessed orders",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ProcessUnprocessedOrders(ctx, mockRepo, wg); (err != nil) != tt.wantErr {
				t.Errorf("ProcessUnprocessedOrders() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_processOrder(t *testing.T) {
	ctx := context.TODO()
	wg := &sync.WaitGroup{}
	util.InitLogger()
	mockRepo := newMockupRepo(t)

	tests := []struct {
		name        string
		orderNumber string
		wantErr     bool
	}{
		{
			name:        "Test process order",
			orderNumber: "34234234",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		wg.Add(1)
		t.Run(tt.name, func(t *testing.T) {
			if err := processOrder(ctx, &tt.orderNumber, mockRepo, wg); (err != nil) != tt.wantErr {
				t.Errorf("processOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func newMockupRepo(t *testing.T) (mockRepo *mocks.MockStorage) {
	ctrl := gomock.NewController(t)

	mockRepo = mocks.NewMockStorage(ctrl)

	mockRepo.EXPECT().StoreGoodsReward(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockRepo.EXPECT().StoreOrder(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockRepo.EXPECT().UpdateOrder(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockRepo.EXPECT().GetOrder(gomock.Any(), gomock.Any()).Return(&domain.OrderRecord{}, nil).AnyTimes()
	mockRepo.EXPECT().GetGoods(gomock.Any()).Return(nil, nil).AnyTimes()

	mockRepo.EXPECT().StoreOrderGoods(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockRepo.EXPECT().GetOrderGoods(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockRepo.EXPECT().GetUnprocessedOrders(gomock.Any()).Return(nil, nil).AnyTimes()

	return
}
