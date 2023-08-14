package domain

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzip(t *testing.T) {
	gzipWriter := GzipResponseWriter{Writer: bytes.NewBuffer([]byte("")), ResponseWriter: http.ResponseWriter(nil)}
	testWritten, err := gzipWriter.Write([]byte("test"))
	require.NoError(t, err)
	assert.Equal(t, 4, testWritten)
}
