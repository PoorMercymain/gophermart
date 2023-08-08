package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PoorMercymain/gophermart/internal/domain"
	"github.com/PoorMercymain/gophermart/pkg/util"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/require"
)

func TestAccrualAddress(t *testing.T) {
	util.InitLogger()

	e := echo.New()

	ts := httptest.NewServer(e)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/test", strings.NewReader(""))
	require.NoError(t, err)

	e.POST("/test", func(ctx echo.Context) error {
		c := ctx.Request()
		addr := c.Context().Value(domain.Key("accrual_address"))
		defer ctx.Response().WriteHeader(http.StatusInternalServerError)
		require.Equal(t, "abc", addr)
		ctx.Response().WriteHeader(http.StatusOK)
		return nil
	}, AddAccrualAddressToCtx("abc"))

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
