package core

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params"
)

type State interface {
	GetState(c common.Address, hash common.Hash) common.Hash
}

func GetBindToken(st State, c *common.Address) (common.Address, bool) {
	storageHash := common.Hash{}
	copy(storageHash[:20], c[:])
	bind := st.GetState(GasTokenBinderAddress, storageHash)
	addressZero := common.Address{}
	b := common.BytesToAddress(bind[:])
	return b, bytes.Equal(b[:], addressZero[:])
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
	config := params.ChainConfig{ChainID: big.NewInt(14298360), HomesteadBlock: big.NewInt(0), DAOForkBlock: nil, DAOForkSupport: false, EIP150Block: big.NewInt(0), EIP150Hash: common.Hash{}, EIP155Block: big.NewInt(0), EIP158Block: big.NewInt(0), ByzantiumBlock: big.NewInt(0), ConstantinopleBlock: big.NewInt(0), PetersburgBlock: big.NewInt(0), IstanbulBlock: big.NewInt(0), MuirGlacierBlock: big.NewInt(0), BerlinBlock: big.NewInt(0), LondonBlock: big.NewInt(0), ArrowGlacierBlock: nil, GrayGlacierBlock: nil, MergeNetsplitBlock: nil, ShanghaiBlock: nil, CancunBlock: nil, TerminalTotalDifficulty: nil, TerminalTotalDifficultyPassed: false, Ethash: nil, Clique: &params.CliqueConfig{Period: 0, Epoch: 30000}}
	config.Clique = &params.CliqueConfig{
		Period: 5,
		Epoch:  30000,
	}
	// Assemble and return the genesis with the precompiles and faucet pre-funded
	routerStorage := make(map[common.Hash]common.Hash, 0)
	routerStorage[common.BytesToHash([]byte{0})] = common.BytesToHash(FactoryAddress[:])
	routerStorage[common.BytesToHash([]byte{1})] = common.BytesToHash(WETHAddress[:])
	faucet := common.HexToAddress("f1658c608708172655a8e70a1624c29f956ee63d")
	a, _ := big.NewInt(0).SetString("100000000000000000000000", 10)
	return &Genesis{
		Config:     &config,
		ExtraData:  hexutil.MustDecode("0x0000000000000000000000000000000000000000000000000000000000000000eb684771f530003b5ea27dcd727d75fcb76658220000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
		GasLimit:   10485760,
		BaseFee:    big.NewInt(params.InitialBaseFee),
		Difficulty: big.NewInt(0),
		Alloc: map[common.Address]GenesisAccount{
			faucet:                {Balance: a},
			FactoryAddress:        {Balance: big.NewInt(0), Code: FactoryCodes},                        // UniswapFactory
			WETHAddress:           {Balance: big.NewInt(0), Code: WETHCodes},                           // WETH
			RouterAddress:         {Balance: big.NewInt(0), Code: RouterCodes, Storage: routerStorage}, // UniswapRouter
			GasTokenBinderAddress: {Balance: big.NewInt(0), Code: GasTokenBinderCodes},                 // GasTokenBinder
		},
	}
}
