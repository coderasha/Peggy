package querier

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/peggy/x/ethbridge/types"
	keeperLib "github.com/cosmos/peggy/x/oracle/keeper"
)

const (
	TestResponseJSON = "{\"id\":\"00x7B95B6EC7EbD73572298cEf32Bb54FA408207359\",\"status\":{\"text\":\"pending\",\"final_claim\":\"\"},\"claims\":[{\"nonce\":0,\"ethereum_sender\":\"0x7B95B6EC7EbD73572298cEf32Bb54FA408207359\",\"cosmos_receiver\":\"cosmos1gn8409qq9hnrxde37kuxwx5hrxpfpv8426szuv\",\"validator_address\":\"cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun\",\"amount\":[{\"denom\":\"ethereum\",\"amount\":\"10\"}]}]}"
)

func TestNewQuerier(t *testing.T) {
	cdc := codec.New()
	ctx, _, keeper, _, _ := keeperLib.CreateTestKeepers(t, 0.7, []int64{3, 3})

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	querier := NewQuerier(keeper, cdc, types.DefaultCodespace)

	//Test wrong paths
	bz, err := querier(ctx, []string{"other"}, query)
	require.NotNil(t, err)
	require.Nil(t, bz)
}

func TestQueryEthProphecy(t *testing.T) {

	cdc := codec.New()
	ctx, _, keeper, _, validatorAddresses := keeperLib.CreateTestKeepers(t, 0.7, []int64{3, 7})
	valAddress := validatorAddresses[0]
	testEthereumAddress := types.NewEthereumAddress(types.TestEthereumAddress)

	initialEthBridgeClaim := types.CreateTestEthClaim(t, valAddress, testEthereumAddress, types.TestCoins)
	oracleClaim, _ := types.CreateOracleClaimFromEthClaim(cdc, initialEthBridgeClaim)
	_, err := keeper.ProcessClaim(ctx, oracleClaim)
	require.Nil(t, err)

	testResponse := types.CreateTestQueryEthProphecyResponse(cdc, t, valAddress)

	//Test query String()
	require.Equal(t, testResponse.String(), TestResponseJSON)

	bz, err2 := cdc.MarshalJSON(types.NewQueryEthProphecyParams(types.TestNonce, testEthereumAddress))
	require.Nil(t, err2)

	query := abci.RequestQuery{
		Path: "/custom/ethbridge/prophecies",
		Data: bz,
	}

	//Test query
	res, err3 := queryEthProphecy(ctx, cdc, query, keeper, types.DefaultCodespace)
	require.Nil(t, err3)

	var ethProphecyResp types.QueryEthProphecyResponse
	err4 := cdc.UnmarshalJSON(res, &ethProphecyResp)
	require.Nil(t, err4)
	require.True(t, reflect.DeepEqual(ethProphecyResp, testResponse))

	// Test error with bad request
	query.Data = bz[:len(bz)-1]

	_, err5 := queryEthProphecy(ctx, cdc, query, keeper, types.DefaultCodespace)
	require.NotNil(t, err5)

	// Test error with nonexistent request
	badEthereumAddress := types.NewEthereumAddress("badEthereumAddress")

	bz2, err6 := cdc.MarshalJSON(types.NewQueryEthProphecyParams(12, badEthereumAddress))
	require.Nil(t, err6)

	query2 := abci.RequestQuery{
		Path: "/custom/oracle/prophecies",
		Data: bz2,
	}

	_, err7 := queryEthProphecy(ctx, cdc, query2, keeper, types.DefaultCodespace)
	require.NotNil(t, err7)

	// Test error with empty address
	emptyEthereumAddress := types.NewEthereumAddress("")

	bz3, err8 := cdc.MarshalJSON(types.NewQueryEthProphecyParams(12, emptyEthereumAddress))
	require.Nil(t, err8)

	query3 := abci.RequestQuery{
		Path: "/custom/oracle/prophecies",
		Data: bz3,
	}

	_, err9 := queryEthProphecy(ctx, cdc, query3, keeper, types.DefaultCodespace)
	require.NotNil(t, err9)
}
