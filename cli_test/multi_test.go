//go:build cli_multi_node_test

package clitest

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Finschia/finschia-sdk/client/flags"
	"github.com/Finschia/finschia-sdk/crypto/keys/ed25519"
	sdk "github.com/Finschia/finschia-sdk/types"

	"github.com/Finschia/finschia/app"
)

func TestMultiValidatorAndSendTokens(t *testing.T) {
	t.Parallel()

	fg := InitFixturesGroup(t)

	fg.FinschiaStartCluster(minGasPrice.String())
	defer fg.Cleanup()

	f := fg.Fixture(0)

	var (
		keyFoo = f.Moniker
	)

	fooAddr := f.KeyAddress(keyFoo)
	f.KeysAdd(keyBaz)
	bazAddr := f.KeyAddress(keyBaz)

	fg.AddFullNode()

	require.NoError(t, fg.Network.WaitForNextBlock())
	{
		fooBal := f.QueryBalances(fooAddr)
		startTokens := sdk.TokensFromConsensusPower(50, sdk.DefaultPowerReduction)
		require.Equal(t, startTokens, fooBal.GetBalances().AmountOf(denom))

		// Send some tokens from one account to the other
		sendTokens := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
		_, err := f.TxSend(keyFoo, bazAddr, sdk.NewCoin(denom, sendTokens), "-y")
		require.NoError(t, err)
		require.NoError(t, fg.Network.WaitForNextBlock())

		// Ensure account balances match expected
		barBal := f.QueryBalances(bazAddr)
		require.Equal(t, sendTokens, barBal.GetBalances().AmountOf(denom))
		fooBal = f.QueryBalances(fooAddr)
		require.Equal(t, startTokens.Sub(sendTokens), fooBal.GetBalances().AmountOf(denom))

		// Test --dry-run
		// _, err = f.TxSend(keyFoo, bazAddr, sdk.NewCoin(denom, sendTokens), "--dry-run")
		// require.NoError(t, err)

		// Test --generate-only
		out, err := f.TxSend(
			fooAddr.String(), bazAddr, sdk.NewCoin(denom, sendTokens), "--generate-only=true",
		)
		require.NoError(t, err)
		msg := UnmarshalTx(f.T, out.Bytes())
		require.NotZero(t, msg.AuthInfo.GetFee().GetGasLimit())
		require.Len(t, msg.GetMsgs(), 1)
		require.Len(t, msg.GetSignatures(), 0)

		// Check state didn't change
		fooBal = f.QueryBalances(fooAddr)
		require.Equal(t, startTokens.Sub(sendTokens), fooBal.GetBalances().AmountOf(denom))

		// test autosequencing
		_, err = f.TxSend(keyFoo, bazAddr, sdk.NewCoin(denom, sendTokens), "-y")
		require.NoError(t, err)
		require.NoError(t, fg.Network.WaitForNextBlock())

		// Ensure account balances match expected
		barBal = f.QueryBalances(bazAddr)
		require.Equal(t, sendTokens.MulRaw(2), barBal.GetBalances().AmountOf(denom))
		fooBal = f.QueryBalances(fooAddr)
		require.Equal(t, startTokens.Sub(sendTokens.MulRaw(2)), fooBal.GetBalances().AmountOf(denom))

		// test note
		_, err = f.TxSend(keyFoo, bazAddr, sdk.NewCoin(denom, sendTokens), fmt.Sprintf("--%s=%s", flags.FlagNote, "testnote"), "-y")
		require.NoError(t, err)
		require.NoError(t, fg.Network.WaitForNextBlock())

		// Ensure account balances match expected
		barBal = f.QueryBalances(bazAddr)
		require.Equal(t, sendTokens.MulRaw(3), barBal.GetBalances().AmountOf(denom))
		fooBal = f.QueryBalances(fooAddr)
		require.Equal(t, startTokens.Sub(sendTokens.MulRaw(3)), fooBal.GetBalances().AmountOf(denom))
	}
}

func TestMultiValidatorAddNodeAndPromoteValidator(t *testing.T) {
	t.Parallel()

	fg := InitFixturesGroup(t)
	fg.FinschiaStartCluster(minGasPrice.String())
	defer fg.Cleanup()

	f1 := fg.Fixture(0)

	f2 := fg.AddFullNode()

	{
		f2.KeysAdd(keyBar)
	}

	barAddr := f2.KeyAddress(keyBar)
	barVal := sdk.ValAddress(barAddr)

	sendTokens := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	{
		_, err := f1.TxSend(f1.Moniker, barAddr, sdk.NewCoin(denom, sendTokens), "-y")
		require.NoError(t, err)
		require.NoError(t, fg.Network.WaitForNextBlock())

		barBal := f2.QueryBalances(barAddr)
		require.Equal(t, sendTokens, barBal.GetBalances().AmountOf(denom))
	}

	newValTokens := sdk.TokensFromConsensusPower(2, sdk.DefaultPowerReduction)
	{
		cdc, _ := app.MakeCodecs()
		pubKeyJSON, err := cdc.MarshalInterfaceJSON(ed25519.GenPrivKey().PubKey())
		require.NoError(t, err)
		consPubKey := string(pubKeyJSON)

		_, err = f2.TxStakingCreateValidator(keyBar, consPubKey, sdk.NewCoin(denom, newValTokens), "-y")
		require.NoError(t, err)
		require.NoError(t, fg.Network.WaitForNextBlock())
	}
	{
		// Ensure funds were deducted properly
		barBal := f2.QueryBalances(barAddr)
		require.Equal(t, sendTokens.Sub(newValTokens), barBal.GetBalances().AmountOf(denom))

		// Ensure that validator state is as expected
		validator := f2.QueryStakingValidator(barVal)
		require.Equal(t, validator.OperatorAddress, barVal.String())
		require.True(sdk.IntEq(t, newValTokens, validator.Tokens))

		// Query delegations to the validator
		validatorDelegations := f2.QueryStakingDelegationsTo(barVal)
		require.Len(t, validatorDelegations.DelegationResponses, 1)
		require.NotZero(t, validatorDelegations.DelegationResponses[0].Delegation.GetShares())
	}
}
