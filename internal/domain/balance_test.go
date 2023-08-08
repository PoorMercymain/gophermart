package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBalance(t *testing.T) {
	testBalance := Balance{Balance: 100, Withdrawn: 1000}

	b := testBalance.Marshal()
	require.NotEmpty(t, b)

	testAccrual := Accrual{Money: 1000}
	b, err := testAccrual.MarshalJSON()
	require.NoError(t, err)
	require.NotEmpty(t, b)

	first, second := getBeforeAndAfterPoint(10001)
	require.Equal(t, 100, first)
	require.Equal(t, 1, second)
}