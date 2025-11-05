package usecase

import (
	"fmt"
	"math/big"
	"strings"
	"sync"
)

var (
	feeNumerator   = big.NewInt(UniswapV2FeeNumerator)
	feeDenominator = big.NewInt(UniswapV2FeeDenominator)
	zeroBig        = big.NewInt(0)

	bigIntPool = sync.Pool{New: func() interface{} { return new(big.Int) }}
)

func (u *EstimateUsecase) parseAmount(amountStr string) (*big.Int, error) {
	amountStr = strings.TrimSpace(amountStr)

	if amountStr == "" {
		return nil, fmt.Errorf("amount cannot be empty")
	}

	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount format: %s", amountStr)
	}

	if amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	return amount, nil
}

func (u *EstimateUsecase) calculateAMMOutput(input, reserveIn, reserveOut *big.Int) (*big.Int, error) {
	if input == nil || reserveIn == nil || reserveOut == nil {
		return nil, fmt.Errorf("nil input/reserves")
	}
	if input.Sign() <= 0 {
		return nil, fmt.Errorf("input amount must be positive")
	}
	if reserveIn.Sign() <= 0 {
		return nil, fmt.Errorf("invalid reserve in: must be positive")
	}
	if reserveOut.Sign() <= 0 {
		return nil, fmt.Errorf("invalid reserve out: must be positive")
	}

	tmpInputWithFee := getTmp()
	tmpNumerator := getTmp()
	tmpReserveInWithFee := getTmp()
	tmpDenominator := getTmp()
	tmpOutput := getTmp()

	defer func() {
		putTmp(tmpInputWithFee)
		putTmp(tmpNumerator)
		putTmp(tmpReserveInWithFee)
		putTmp(tmpDenominator)
		putTmp(tmpOutput)
	}()

	tmpInputWithFee.Mul(input, feeNumerator)

	tmpNumerator.Mul(tmpInputWithFee, reserveOut)

	tmpReserveInWithFee.Mul(reserveIn, feeDenominator)

	tmpDenominator.Add(tmpReserveInWithFee, tmpInputWithFee)

	tmpOutput.Div(tmpNumerator, tmpDenominator)

	out := new(big.Int).Set(tmpOutput)
	if out.Sign() <= 0 {
		return out, nil
	}
	return out, nil
}

func getTmp() *big.Int {
	return bigIntPool.Get().(*big.Int)
}

func putTmp(x *big.Int) {
	x.SetInt64(0)
	bigIntPool.Put(x)
}
