package handler

import (
	"github.com/PoorMercymain/gophermart/internal/accrual/domain"
	"github.com/PoorMercymain/gophermart/internal/accrual/interfaces/mocks"
	"github.com/PoorMercymain/gophermart/internal/accrual/middleware"

	"github.com/PoorMercymain/gophermart/pkg/util"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testRouter(t *testing.T) *echo.Echo {

	e := echo.New()

	util.InitLogger()

	ctrl := gomock.NewController(t)

	mockRepo := mocks.NewMockStorage(ctrl)

	mockRepo.EXPECT().StoreGoodsReward(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockRepo.EXPECT().StoreOrder(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockRepo.EXPECT().UpdateOrder(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockRepo.EXPECT().GetOrder(gomock.Any(), gomock.Any()).Return(&domain.OrderRecord{}, nil).AnyTimes()
	mockRepo.EXPECT().GetGoods(gomock.Any()).Return(nil, nil).AnyTimes()

	mockRepo.EXPECT().StoreOrderGoods(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockRepo.EXPECT().GetOrderGoods(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockRepo.EXPECT().GetUnprocessedOrders(gomock.Any()).Return(nil, nil).AnyTimes()

	sh := NewStorageHandler(mockRepo)

	e.Use(middleware.UseGzipReader())

	e.GET("/api/orders/:number", sh.ProcessGetOrdersRequest)
	e.POST("/api/orders", sh.ProcessPostOrdersRequest)
	e.POST("/api/goods", sh.ProcessPostGoodsRequest)

	return e
}

func request(t *testing.T, ts *httptest.Server, code int, method, body, endpoint string) *http.Response {
	req, err := http.NewRequest(method, ts.URL+endpoint, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.Client().Do(req)
	if err != http.ErrUseLastResponse {
		require.NoError(t, err)
	}
	defer resp.Body.Close()

	require.Equal(t, code, resp.StatusCode)

	return resp
}

func TestRouter(t *testing.T) {
	ts := httptest.NewServer(testRouter(t))

	defer ts.Close()

	var testTable = []struct {
		endpoint string
		method   string
		code     int
		body     string
	}{
		{"/api/orders/1234456791", http.MethodGet, http.StatusNoContent, ""},
		{"/api/orders", http.MethodGet, http.StatusMethodNotAllowed, ""},
		{"/api/orders/0927", http.MethodGet, http.StatusOK, ""},

		{"/api/orders", http.MethodPost, http.StatusBadRequest, "{}"},
		{"/api/orders", http.MethodPost, http.StatusBadRequest, "{\"order\":\"1234456791\"}"},

		//{"/api/goods", http.MethodPost, http.StatusAccepted, "573956"},
	}

	for _, testCase := range testTable {
		resp := request(t, ts, testCase.code, testCase.method, testCase.body, testCase.endpoint)
		resp.Body.Close()
	}
}
