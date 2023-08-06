package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/PoorMercymain/gophermart/internal/domain"
	"github.com/PoorMercymain/gophermart/internal/domain/mocks"
	"github.com/PoorMercymain/gophermart/internal/middleware"
	"github.com/PoorMercymain/gophermart/internal/service"
	"github.com/PoorMercymain/gophermart/pkg/util"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/require"
)

func testRouter(t *testing.T) *echo.Echo {
	e := echo.New()
	util.InitLogger()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockRepo.EXPECT().GetPasswordHash(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
	mockRepo.EXPECT().Register(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockRepo.EXPECT().AddOrder(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockRepo.EXPECT().ReadOrders(gomock.Any()).Return(nil, nil).AnyTimes()
	mockRepo.EXPECT().ReadBalance(gomock.Any()).Return(domain.Balance{}, nil).AnyTimes()
	mockRepo.EXPECT().AddWithdrawal(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockRepo.EXPECT().ReadWithdrawals(gomock.Any()).Return(nil, nil).AnyTimes()

	us := service.NewUser(mockRepo)
	uh := NewUser(us)

	var wg sync.WaitGroup
	e.POST("/api/user/register", uh.Register, middleware.UseGzipReader())
	e.POST("/api/user/login", uh.Authenticate, middleware.UseGzipReader())
	e.POST("/api/user/orders", uh.AddOrder(&wg), middleware.UseGzipReader(), middleware.AddAccrualAddressToCtx(""), middleware.AddTestingToCtx())
	e.GET("/api/user/orders", uh.ReadOrders, middleware.UseGzipReader())
	e.GET("/api/user/balance", uh.ReadBalance, middleware.UseGzipReader())
	e.POST("/api/user/balance/withdraw", uh.AddWithdrawal, middleware.UseGzipReader())
	e.GET("/api/user/withdrawals", uh.ReadWithdrawals, middleware.UseGzipReader())

	return e
}

func request(t *testing.T, ts *httptest.Server, code int, method, body, endpoint string) *http.Response {
	req, err := http.NewRequest(method, ts.URL+endpoint, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	if endpoint == "/api/user/orders" {
		req.Header.Set("Content-Type", "text/plain")
	}

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
		{"/api/user/register", http.MethodPost, http.StatusOK, "{\"login\":\"test\",\"password\":\"test\"}"},
		{"/api/user/login", http.MethodPost, http.StatusUnauthorized, "{\"login\":\"test\",\"password\":\"testing\"}"},
		{"/api/user/orders", http.MethodPost, http.StatusAccepted, "573956"},
		{"/api/user/orders", http.MethodGet, http.StatusNoContent, ""},
		{"/api/user/balance", http.MethodGet, http.StatusOK, ""},
		{"/api/user/balance/withdraw", http.MethodPost, http.StatusOK, "{\"order\": \"573956\", \"sum\": 0}"},
		{"/api/user/withdrawals", http.MethodGet, http.StatusNoContent, ""},
	}

	for _, testCase := range testTable {
		resp := request(t, ts, testCase.code, testCase.method, testCase.body, testCase.endpoint)
		resp.Body.Close()
	}
}
