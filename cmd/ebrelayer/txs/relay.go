package txs

// ------------------------------------------------------------
//	Relay : Builds and encodes EthBridgeClaim Msgs with the
//  	specified variables, before presenting the unsigned
//      transaction to validators for optional signing.
//      Once signed, the data packets are sent as transactions
//      on the Cosmos Bridge.
// ------------------------------------------------------------

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	amino "github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"

	"github.com/cosmos/peggy/x/ethbridge"
	"github.com/cosmos/peggy/x/ethbridge/types"
)

// RelayEvent : RelayEvent applies validator's signature to an EthBridgeClaim message
//		containing information about an event on the Ethereum blockchain before sending
//		it to the Bridge blockchain. For this relay, the chain id (chainID) and codec
//		(cdc) of the Bridge blockchain are required.
//
func RelayEvent(chainID string, cdc *amino.Codec, validatorAddress sdk.ValAddress, moniker string, passphrase string, claim *types.EthBridgeClaim) error {

	cliCtx := context.NewCLIContext().
		WithCodec(cdc).
		WithAccountDecoder(cdc).
		WithFromAddress(sdk.AccAddress(validatorAddress)).
		WithFromName(moniker)

	cliCtx.SkipConfirm = true

	txBldr := authtxb.NewTxBuilderFromCLI().
		WithTxEncoder(utils.GetTxEncoder(cdc)).
		WithChainID(chainID)

	err := cliCtx.EnsureAccountExistsFromAddr(sdk.AccAddress(claim.ValidatorAddress))
	if err != nil {
		return err
	}

	msg := ethbridge.NewMsgCreateEthBridgeClaim(*claim)

	err = msg.ValidateBasic()
	if err != nil {
		return err
	}

	cliCtx.PrintResponse = true

	// Prepare tx
	txBldr, err = utils.PrepareTxBuilder(txBldr, cliCtx)
	if err != nil {
		return err
	}

	// Build and sign the transaction
	txBytes, err := txBldr.BuildAndSign(moniker, passphrase, []sdk.Msg{msg})
	if err != nil {
		return err
	}

	// Broadcast to a Tendermint node
	res, err := cliCtx.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	cliCtx.PrintOutput(res)
	return nil
}
