package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PoorMercymain/gophermart/pkg/util"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/require"
)

func TestUseGzipReader(t *testing.T) {
	util.InitLogger()

	e := echo.New()

	ts := httptest.NewServer(e)
	defer ts.Close()

	buf := bytes.NewBuffer([]byte(""))
	w := gzip.NewWriter(buf)
	w.Write([]byte("12345"))
	w.Close()
	r := bytes.NewBuffer(buf.Bytes())

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/test", r)
	require.NoError(t, err)

	req.Header.Set("Content-Encoding", "gzip")

	req2, err := http.NewRequest(http.MethodPost, ts.URL+"/test", r)
	require.NoError(t, err)

	e.POST("/test", func(ctx echo.Context) error {
		c := ctx.Request()
		if c.Header.Get("Content-Encoding") != "gzip" {
			util.GetLogger().Infoln("Content-Encoding is not gzip")
			ctx.Response().WriteHeader(http.StatusBadRequest)
			return nil
		} else {
			var b []byte
			b, err = io.ReadAll(ctx.Request().Body)
			ctx.Request().Body.Close()
			util.GetLogger().Infoln(err)
			require.NoError(t, err)
			util.GetLogger().Infoln(string(b))
			ctx.Response().WriteHeader(http.StatusOK)
			return nil
		}
	}, UseGzipReader())

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = ts.Client().Do(req2)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
