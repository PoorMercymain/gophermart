package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/PoorMercymain/gophermart/internal/domain"
	"github.com/PoorMercymain/gophermart/internal/domain/mocks"
	"github.com/PoorMercymain/gophermart/internal/middleware"
	"github.com/PoorMercymain/gophermart/internal/service"
	"github.com/PoorMercymain/gophermart/pkg/util"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func testRouter(t *testing.T) *echo.Echo {
	e := echo.New()
	util.InitLogger()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)

	testHash, err := bcrypt.GenerateFromPassword([]byte("test"), bcrypt.DefaultCost)
	require.NoError(t, err)
	testHashStr := string(testHash)
	testDomainOrder := make([]domain.Order, 0)
	tdo := domain.Order{
		Number:           "573956",
		Status:           "PROCESSED",
		Accrual:          domain.Accrual{Money: 1000},
		UploadedAt:       time.Now(),
		UploadedAtString: time.Now().Format(time.RFC3339),
	}
	testDomainOrder = append(testDomainOrder, tdo)

	mockRepo.EXPECT().GetPasswordHash(gomock.Any(), gomock.Any()).Return(testHashStr, nil).AnyTimes()
	mockRepo.EXPECT().Register(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockRepo.EXPECT().AddOrder(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockRepo.EXPECT().ReadOrders(gomock.Any()).Return(nil, nil).MaxTimes(1)
	mockRepo.EXPECT().ReadOrders(gomock.Any()).Return(testDomainOrder, nil).AnyTimes()
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

func request(t *testing.T, ts *httptest.Server, code int, method, content, body, endpoint string) *http.Response {
	req, err := http.NewRequest(method, ts.URL+endpoint, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", content)

	resp, err := ts.Client().Do(req)
	if err != http.ErrUseLastResponse {
		require.NoError(t, err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	util.GetLogger().Infoln(string(b))

	require.Equal(t, code, resp.StatusCode)

	return resp
}

func TestRouter(t *testing.T) {
	ts := httptest.NewServer(testRouter(t))

	defer ts.Close()

	var testTable = []struct {
		endpoint string
		method   string
		content  string
		code     int
		body     string
	}{
		{"/api/user/register", http.MethodPost, "application/json", http.StatusOK, "{\"login\":\"test\",\"password\":\"test\"}"},
		{"/api/user/register", http.MethodPost, "text/plain", http.StatusBadRequest, "{\"login\":\"test\",\"password\":\"test\"}"},
		{"/api/user/register", http.MethodPost, "application/json", http.StatusBadRequest, "{\"1\":\"2\",\"login\":\"test\",\"password\":\"test\"}"},
		{"/api/user/register", http.MethodPost, "application/json", http.StatusBadRequest, "{\"login\":\"test\"}"},
		{"/api/user/register", http.MethodPost, "application/json", http.StatusBadRequest, "{\"login\":\"test\",\"login\":\"test\",\"password\":\"test\"}"},
		{"/api/user/register", http.MethodPost, "application/json", http.StatusBadRequest, "{\"login\":\"test\"\"password\":\"test\"}"},
		{"/api/user/login", http.MethodPost, "application/json", http.StatusOK, "{\"login\":\"test\",\"password\":\"test\"}"},
		{"/api/user/login", http.MethodPost, "application/json", http.StatusUnauthorized, "{\"login\":\"test\",\"password\":\"testing\"}"},
		{"/api/user/login", http.MethodPost, "text/plain", http.StatusBadRequest, "{\"login\":\"test\",\"password\":\"testing\"}"},
		{"/api/user/login", http.MethodPost, "application/json", http.StatusBadRequest, "{\"login\":\"test\",\"password\":\"testing\",\"password\":\"testing\"}"},
		{"/api/user/login", http.MethodPost, "application/json", http.StatusBadRequest, "{\"login\":\"test\",\"password\":\"testing\""},
		{"/api/user/login", http.MethodPost, "application/json", http.StatusBadRequest, "{\"login\":\"test\","},
		{"/api/user/orders", http.MethodPost, "text/plain", http.StatusAccepted, "573956"},
		{"/api/user/orders", http.MethodPost, "text/plain", http.StatusUnprocessableEntity, "12345"},
		{"/api/user/orders", http.MethodPost, "text/plain", http.StatusBadRequest, "abc12345"},
		{"/api/user/orders", http.MethodPost, "application/json", http.StatusBadRequest, "573956"},
		{"/api/user/orders", http.MethodGet, "", http.StatusNoContent, ""},
		{"/api/user/orders", http.MethodGet, "", http.StatusOK, ""},
		{"/api/user/balance", http.MethodGet, "", http.StatusOK, ""},
		{"/api/user/balance/withdraw", http.MethodPost, "application/json", http.StatusOK, "{\"order\": \"573956\", \"sum\": 0}"},
		{"/api/user/balance/withdraw", http.MethodPost, "text/plain", http.StatusBadRequest, "{\"order\": \"573956\", \"sum\": 0}"},
		{"/api/user/balance/withdraw", http.MethodPost, "application/json", http.StatusBadRequest, "{\"order\": \"573956\""},
		{"/api/user/balance/withdraw", http.MethodPost, "application/json", http.StatusUnprocessableEntity, "{\"order\": \"12345\", \"sum\":0}"},
		{"/api/user/balance/withdraw", http.MethodPost, "application/json", http.StatusBadRequest, "{\"order\": \"12345abc\", \"sum\":0}"},
		{"/api/user/withdrawals", http.MethodGet, "", http.StatusNoContent, ""},
	}

	for _, testCase := range testTable {
		util.GetLogger().Infoln("called", testCase.endpoint)
		resp := request(t, ts, testCase.code, testCase.method, testCase.content, testCase.body, testCase.endpoint)
		resp.Body.Close()
	}
}

func TestUtils(t *testing.T) {
	r, err := http.NewRequest("POST", "", bytes.NewReader([]byte("")))
	require.NoError(t, err)
	r.Header.Set("Content-Type", "application/json")
	isCorrect := IsJSONContentTypeCorrect(r)
	assert.Equal(t, true, isCorrect)
	r.Header.Set("Content-Type", "text/plain")
	isCorrect = IsJSONContentTypeCorrect(r)
	assert.Equal(t, false, isCorrect)
	isCorrect = IsPlaintextContentTypeCorrect(r)
	assert.Equal(t, true, isCorrect)
	r.Header.Set("Content-Type", "application/json")
	isCorrect = IsPlaintextContentTypeCorrect(r)
	assert.Equal(t, false, isCorrect)
	testJWTString, err := CreateJWTString("abcd")
	require.NoError(t, err)
	assert.NotEmpty(t, testJWTString)
}

func TestStartup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	us := service.NewUser(mockRepo)
	uh := NewUser(us)

	mockRepo.EXPECT().GetUnprocessedBatch(gomock.Any(), gomock.Any()).Return(make([]domain.AccrualOrderWithUsername, 0), nil).AnyTimes()

	var wg sync.WaitGroup
	err := uh.HandleStartup("", &wg)
	assert.NoError(t, err)
}
