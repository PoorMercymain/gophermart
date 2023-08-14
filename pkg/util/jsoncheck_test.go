package util

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJsonCheck(t *testing.T) {
	err := dupErr(make([]string, 0))
	require.ErrorIs(t, err, ErrDuplicate)

	d := json.NewDecoder(strings.NewReader("{\"test\":100}"))
	err = CheckDuplicatesInJSON(d, nil)
	require.NoError(t, err)

	d = json.NewDecoder(strings.NewReader("{\"test\":100, \"test\":200}"))
	err = CheckDuplicatesInJSON(d, nil)
	require.Error(t, err)
}
