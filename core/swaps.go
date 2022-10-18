package core

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

type State interface {
	GetState(c common.Address, hash common.Hash) common.Hash
}

func GetGasToken(st State, c *common.Address) (common.Address, bool) {
	storageHash := common.Hash{}
	copy(storageHash[:20], c[:])
	bind := st.GetState(GasTokenBinderAddress, storageHash)
	addressZero := common.Address{}
	b := common.BytesToAddress(bind[:])
	return b, !bytes.Equal(b[:], addressZero[:])
}

func GetTokenBalanceOf(evm *vm.EVM, token common.Address, caller common.Address) (*big.Int, error) {
	balanceOfABI := `[{"constant":true,"inputs":[{"internalType":"address","name":"","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}]`
	var (
		balanceOfFunc abi.ABI
		err           error
	)
	balanceOfFunc, err = abi.JSON(bytes.NewReader([]byte(balanceOfABI)))
	if err != nil {
		panic(err)
	}
	input, err := balanceOfFunc.Pack("balanceOf", caller)
	if err != nil {
		return big.NewInt(0), err
	}
	snap := evm.StateDB.Snapshot()
	defer func() {
		evm.StateDB.RevertToSnapshot(snap)
	}()
	ret, _, err := evm.SystemStaticCall(vm.AccountRef(caller), token, input, uint64(MaxBalanceOfGas))
	if err != nil {
		return big.NewInt(0), err
	}
	uintType, err := abi.NewType("uint256", "", nil)
	if err != nil {
		panic(err)
	}
	args := abi.Arguments{
		abi.Argument{
			Type: uintType,
		},
	}
	data, err := args.Unpack(ret)
	if err != nil {
		return big.NewInt(0), err
	}
	if len(data) == 0 {
		return big.NewInt(0), fmt.Errorf("invalid unpacked data")
	}
	return data[0].(*big.Int), nil
}

func GetAmountsIn(evm *vm.EVM, token common.Address, caller common.Address, value *big.Int) (*big.Int, error) {
	getAmounsInABI := `[{"inputs":[{"internalType":"uint256","name":"amountOut","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"}],"name":"getAmountsIn","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"view","type":"function"}]`
	var (
		getAmounsInFunc abi.ABI
		err             error
	)
	getAmounsInFunc, err = abi.JSON(bytes.NewReader([]byte(getAmounsInABI)))
	if err != nil {
		panic(err)
	}
	path := make([]common.Address, 0)
	path = append(path, token, WETHAddress)
	input, err := getAmounsInFunc.Pack("getAmountsIn", value, path)
	if err != nil {
		return big.NewInt(0), err
	}
	snap := evm.StateDB.Snapshot()
	defer func() {
		evm.StateDB.RevertToSnapshot(snap)
	}()
	ret, _, err := evm.SystemStaticCall(vm.AccountRef(caller), RouterAddress, input, uint64(MaxGetAmountsInGas))
	if err != nil {
		return big.NewInt(0), err
	}
	uintArrayType, err := abi.NewType("uint256[]", "", nil)
	if err != nil {
		panic(err)
	}
	args := abi.Arguments{
		abi.Argument{
			Type: uintArrayType,
		},
	}
	data, err := args.Unpack(ret)
	if err != nil {
		return big.NewInt(0), err
	}
	if len(data) == 0 {
		return big.NewInt(0), fmt.Errorf("invalid unpacked data")
	}
	if v, ok := data[0].([]*big.Int); ok {
		if len(v) == 0 {
			return big.NewInt(0), fmt.Errorf("invalid unpacked data")
		}
		return v[0], nil
	} else {
		return big.NewInt(0), fmt.Errorf("invalid unpacked data")
	}
}

func NewETHSwapData(amountOut *big.Int, amountInMax *big.Int, token common.Address, to common.Address, deadline *big.Int) (data []byte) {
	swapABI := `[{"inputs":[{"internalType":"uint256","name":"amountOut","type":"uint256"},{"internalType":"uint256","name":"amountInMax","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapTokensForExactETH","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"nonpayable","type":"function"}]`
	var (
		swapFunc abi.ABI
		err      error
	)
	swapFunc, err = abi.JSON(bytes.NewReader([]byte(swapABI)))
	if err != nil {
		panic(err)
	}
	path := make([]common.Address, 0)
	path = append(path, token, WETHAddress)
	data, err = swapFunc.Pack("swapTokensForExactETH", amountOut, amountInMax, path, to, deadline)
	if err != nil {
		panic(err)
	}
	return data
}

func DefaulSwapDevGenesisBlock() *Genesis {
	// Override the default period to the user requested one
	config := params.ChainConfig{ChainID: big.NewInt(1130), HomesteadBlock: big.NewInt(0), DAOForkBlock: nil, DAOForkSupport: false, EIP150Block: big.NewInt(0), EIP150Hash: common.Hash{}, EIP155Block: big.NewInt(0), EIP158Block: big.NewInt(0), ByzantiumBlock: big.NewInt(0), ConstantinopleBlock: big.NewInt(0), PetersburgBlock: big.NewInt(0), IstanbulBlock: big.NewInt(0), MuirGlacierBlock: big.NewInt(0), BerlinBlock: big.NewInt(0), LondonBlock: big.NewInt(0), ArrowGlacierBlock: nil, GrayGlacierBlock: nil, MergeNetsplitBlock: nil, ShanghaiBlock: nil, CancunBlock: nil, TerminalTotalDifficulty: nil, TerminalTotalDifficultyPassed: false, Ethash: nil, Clique: &params.CliqueConfig{Period: 0, Epoch: 30000}}
	config.Clique = &params.CliqueConfig{
		Period: 3,
		Epoch:  30000,
	}
	routerStorage := map[common.Hash]common.Hash{
		common.BytesToHash([]byte{0}): common.BytesToHash(FactoryAddress[:]),
		common.BytesToHash([]byte{1}): common.BytesToHash(WETHAddress[:]),
	}
	faucet := common.HexToAddress("f1658c608708172655a8e70a1624c29f956ee63d")
	faucet1 := common.HexToAddress("99F5a620384A5a530320Fc0Be2af3b69D763ED4f")
	faucet2 := common.HexToAddress("D83c37edF25DCE7FD965a8E7aA0f22F5bbfa91Ca")
	a, _ := big.NewInt(0).SetString("1000000000000000000000000000", 10)
	return &Genesis{
		Config:     &config,
		ExtraData:  hexutil.MustDecode("0x0000000000000000000000000000000000000000000000000000000000000000AcD534B544e5E83E09E1394CecFBfAF6Fca61E0A0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
		GasLimit:   10485760,
		BaseFee:    big.NewInt(params.InitialBaseFee),
		Difficulty: big.NewInt(0),
		Alloc: map[common.Address]GenesisAccount{
			faucet:                {Balance: a},
			faucet1:               {Balance: a},
			faucet2:               {Balance: a},
			FactoryAddress:        {Balance: big.NewInt(0), Code: FactoryCodes},                        // UniswapFactory
			WETHAddress:           {Balance: big.NewInt(0), Code: WETHCodes},                           // WETH
			RouterAddress:         {Balance: big.NewInt(0), Code: RouterCodes, Storage: routerStorage}, // UniswapRouter
			GasTokenBinderAddress: {Balance: big.NewInt(0), Code: GasTokenBinderCodes},                 // GasTokenBinder
		},
	}
}
