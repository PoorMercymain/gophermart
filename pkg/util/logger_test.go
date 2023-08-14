package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	err := LogInfoln()
	require.Error(t, err)

	logger := GetLogger()
	require.Empty(t, logger)

	err = InitLogger()
	require.NoError(t, err)

	logger = GetLogger()
	require.NotEmpty(t, logger)

	err = LogInfoln()
	require.NoError(t, err)
}
