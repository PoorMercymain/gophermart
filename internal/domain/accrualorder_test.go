package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccrualAmount(t *testing.T) {
	testAccrualAmount := AccrualAmount{}

	err := testAccrualAmount.UnmarshalJSON([]byte("100.5"))
	require.NoError(t, err)
	require.NotEmpty(t, testAccrualAmount.Accrual)

	err = testAccrualAmount.UnmarshalJSON([]byte("asdf"))
	require.Error(t, err)
}
