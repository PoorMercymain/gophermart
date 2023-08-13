package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithdrawal(t *testing.T) {
	testWithdrawalAmount := WithdrawalAmount{Withdrawal: 100}

	marshaled, err := testWithdrawalAmount.MarshalJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, marshaled)

	err = testWithdrawalAmount.UnmarshalJSON(marshaled)
	require.NoError(t, err)

	testWithdrawalAmount = WithdrawalAmount{Withdrawal: 1}

	marshaled, err = testWithdrawalAmount.MarshalJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, marshaled)

	err = testWithdrawalAmount.UnmarshalJSON(marshaled)
	require.NoError(t, err)

	testWithdrawalAmount = WithdrawalAmount{Withdrawal: 110}

	marshaled, err = testWithdrawalAmount.MarshalJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, marshaled)

	err = testWithdrawalAmount.UnmarshalJSON(marshaled)
	require.NoError(t, err)

	err = testWithdrawalAmount.UnmarshalJSON(nil)
	require.Error(t, err)

}
