package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PoorMercymain/gophermart/internal/domain"
	"github.com/PoorMercymain/gophermart/internal/domain/mocks"
	"github.com/PoorMercymain/gophermart/internal/handler"
	"github.com/PoorMercymain/gophermart/pkg/util"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestGetUser(t *testing.T) {
	util.InitLogger()

	usr := domain.User{Login: "test", Password: "test"}

	jwtStr, err := handler.CreateJWTString(usr.Login + " " + usr.Password)
	require.NoError(t, err)

	u, err := GetUserFromJWT(jwtStr)
	require.NoError(t, err)
	require.Equal(t, usr, u)

	_, err = GetUserFromJWT("abcde")
	require.Error(t, err)
}

func TestAuth(t *testing.T) {
	util.InitLogger()

	util.GetLogger().Infoln("2")
	e := echo.New()

	ts := httptest.NewServer(e)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/test", strings.NewReader(""))
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)

	testHash, err := bcrypt.GenerateFromPassword([]byte("test"), bcrypt.DefaultCost)
	require.NoError(t, err)
	testHashStr := string(testHash)

	usr := domain.User{Login: "test", Password: testHashStr}
	jwtStr, err := handler.CreateJWTString(usr.Login + " " + usr.Password)
	require.NoError(t, err)

	req.Header.Set("Authorization", jwtStr)

	mockRepo.EXPECT().GetPasswordHash(gomock.Any(), gomock.Any()).Return(testHashStr, nil)

	e.POST("/test", func(ctx echo.Context) error {
		ctx.Response().WriteHeader(http.StatusOK)
		return nil
	}, CheckAuth(mockRepo))

	resp, err := ts.Client().Do(req)
	util.GetLogger().Infoln("")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}
