package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank"

	"github.com/certikfoundation/shentu/x/auth/vesting"
	"github.com/certikfoundation/shentu/x/bank/internal/types"
)

// NewHandler returns a handler for "auth" type messages.
func NewHandler(k Keeper, ak types.AccountKeeper) sdk.Handler {
	cosmosHandler := bank.NewHandler(k)
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case types.MsgLockedSend:
			return handleMsgLockedSend(ctx, k, ak, msg)
		default:
			return cosmosHandler(ctx, msg)
		}
	}
}

func handleMsgLockedSend(ctx sdk.Context, k Keeper, ak types.AccountKeeper, msg types.MsgLockedSend) (*sdk.Result, error) {
	// preliminary checks
	acc := ak.GetAccount(ctx, msg.From)
	if acc == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", msg.From)
	}

	acc = ak.GetAccount(ctx, msg.To)
	if acc == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", msg.To)
	}
	toAcc, ok := acc.(*vesting.ManualVestingAccount)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "receiver account does not appear to be a ManualVestingAccount")
	}

	//TODO: event?

	// subtraction from sender account (as normally done)
	_, err := k.SubtractCoins(ctx, msg.From, msg.Amount)
	if err != nil {
		return nil, err
	}

	// add to receiver account as normally done
	// but make the added amount vesting (OV := Vesting + Vested)
	toAcc.OriginalVesting = toAcc.OriginalVesting.Add(msg.Amount...)
	ak.SetAccount(ctx, toAcc)

	_, err = k.AddCoins(ctx, msg.To, msg.Amount)
	if err != nil {
		return nil, err
	}

	return &sdk.Result{}, nil
}