package eths

import (
	"context"
	"encoding/json"
	"fmt"
	"go-echo/configs"
	"go-echo/dbs"
	"go-echo/utils"
	"math/big"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// CountStorage ...
var CountStorage map[int64]*big.Int

// TranDetail ...
type TranDetail struct {
	From common.Address `json:"from"`
	To   common.Address `json:"to"`
	// Value *hexutil.Big   `json:"value"`
}

// NewAcc ...
func NewAcc(pass, connstr string) (string, error) {
	cli, err := rpc.Dial(connstr)
	if err != nil {
		fmt.Println("failed to connect to geth", err)
		return "", err
	}
	defer cli.Close()
	var account string
	err = cli.Call(&account, "personal_newAccount", pass)
	if err != nil {
		fmt.Println("failed to connect to personal_newAccount", err)
		return "", err
	}
	fmt.Println("account build successfully")
	return account, err
}

// Upload ...
func Upload(from, pass, hash, data string, price, weight int64) error {
	cli, err := ethclient.Dial(configs.Config.Eth.Connstr)
	if err != nil {
		fmt.Println("failed to ethclient.Dial", err)
		return err
	}
	instance, err := NewPxa(common.HexToAddress(configs.Config.Eth.PxaAddr), cli)
	if err != nil {
		fmt.Println("failed to eths.NewPxa", err)
		return err
	}
	// 设置签名, owner的keyStore文件
	// 需要获得文件名字
	fileName, err := utils.GetFileName(string([]rune(from)[2:]), configs.Config.Eth.Keydir)

	file, err := os.Open(configs.Config.Eth.Keydir + "/" + fileName)
	if err != nil {
		fmt.Println("failed to os.Open", err)
		return err
	}
	auth, err := bind.NewTransactor(file, pass)
	if err != nil {
		fmt.Println("failed to bind.NewTransactor", err)
		return err
	}
	// string -> [32]byte
	_, err = instance.Mint(auth, common.HexToHash(hash), big.NewInt(price), big.NewInt(weight), data)
	if err != nil {
		fmt.Println("failed to instance.Mint", err)
		return err
	}
	fmt.Printf("the account: %s Mint success...\n", from)
	return nil
}

// EventSubscribeTest ...
func EventSubscribeTest(connstr, contractAddr string) error {
	// 1.连接ws://localhost:8546
	cli, err := ethclient.Dial(connstr)
	if err != nil {
		fmt.Println("failed to ethclient.Dial", err)
		return err
	}
	// 2. 合约地址处理
	cAddress := common.HexToAddress(contractAddr)
	newAssetHash := crypto.Keccak256Hash([]byte("onNewAsset(bytes32,address,uint256)"))
	// 3. 过滤处理
	query := ethereum.FilterQuery{
		Addresses: []common.Address{cAddress},
		Topics:    [][]common.Hash{{newAssetHash}},
	}
	// 4. 通道
	pxaLogs := make(chan types.Log)
	// 5. 订阅
	sub, err := cli.SubscribeFilterLogs(context.Background(), query, pxaLogs)
	if err != nil {
		fmt.Println("failed to cli.SubscribeFilterLogs", err)
		return err
	}
	// 6. 订阅返回处理
	fmt.Println("starting operate sub...")
	for {
		select {
		case err := <-sub.Err():
			fmt.Println("get sub err", err)
		case vLog := <-pxaLogs:
			data, err := vLog.MarshalJSON()
			fmt.Println(string(data), err)
			ParseMintEventDb([]byte(common.Bytes2Hex(vLog.Data)))
		}
	}
}

// EthSplitAsset ...
func EthSplitAsset(fundation, pass, buyer string, tokenID, weight int64) error {
	cli, err := ethclient.Dial(configs.Config.Eth.Connstr)
	if err != nil {
		fmt.Println("failed to ethclient.Dial", err)
		return err
	}
	instance, err := NewPxa(common.HexToAddress(configs.Config.Eth.PxaAddr), cli)
	if err != nil {
		fmt.Println("failed to eths.NewPxa", err)
		return err
	}
	// 设置签名, owner的keyStore文件
	// 需要获得文件名字
	fileName, err := utils.GetFileName(string([]rune(fundation)[2:]), configs.Config.Eth.Keydir)
	file, err := os.Open(configs.Config.Eth.Keydir + "/" + fileName)
	if err != nil {
		fmt.Println("failed to os.Open", err)
		return err
	}
	auth, err := bind.NewTransactor(file, pass)
	if err != nil {
		fmt.Println("failed to bind.NewTransactor", err)
		return err
	}
	// string -> [32]byte
	// SplitAsset(opts *bind.TransactOpts, _tokenId *big.Int, _weight *big.Int, _buyer common.Address)
	// fmt.Printf("tokenID: %d, weight: %d, buyer:%v auth:%v\n", big.NewInt(tokenID), big.NewInt(weight), buyer, fundation)

	// 分割新的资产, 添加事件, 将新的资产存储content文本中
	_, err = instance.SplitAsset(auth, big.NewInt(tokenID), big.NewInt(weight), common.HexToAddress(buyer))

	if err != nil {
		fmt.Println("failed to SplitAsset", err)
		return err
	}
	fmt.Printf("the account: %s SplitAsset success...\n", fundation)
	// 将新的资产存入数据库中, 获取新的asset, token_id = asset.length-1
	return nil
}

// EthErc20Transfer ...
func EthErc20Transfer(from, pass, seller string, num int64) error {
	cli, err := ethclient.Dial(configs.Config.Eth.Connstr)
	if err != nil {
		fmt.Println("failed to ethclient.Dial", err)
		return err
	}
	instance, err := NewPxc(common.HexToAddress(configs.Config.Eth.PxcAddr), cli)
	if err != nil {
		fmt.Println("failed to eths.NewPxc", err)
		return err
	}
	// 设置签名, owner的keyStore文件
	// 需要获得文件名字
	fileName, err := utils.GetFileName(string([]rune(from)[2:]), configs.Config.Eth.Keydir)
	file, err := os.Open(configs.Config.Eth.Keydir + "/" + fileName)
	if err != nil {
		fmt.Println("failed to os.Open", err)
		return err
	}
	auth, err := bind.NewTransactor(file, pass)
	if err != nil {
		fmt.Println("failed to bind.NewTransactor", err)
		return err
	}
	// string -> [32]byte
	// Transfer(opts *bind.TransactOpts, _to common.Address, _value *big.Int)
	_, err = instance.Transfer(auth, common.HexToAddress(seller), big.NewInt(num))
	if err != nil {
		fmt.Println("failed to Transfer", err)
		return err
	}
	fmt.Printf("the account: %s Transfer success...\n", from)
	return nil
}

// EtherTransfer ...
func EtherTransfer(from, newAcc string) (string, error) {
	cli, err := rpc.Dial(configs.Config.Eth.Connstr)
	if err != nil {
		fmt.Println("failed to ethclient.Dial", err)
		return "", err
	}

	defer cli.Close()
	var transcationHash string
	fmt.Printf("from: %s, to：%s\n", from, newAcc)
	t := &TranDetail{common.HexToAddress(from), common.HexToAddress(newAcc)}
	data, err := json.Marshal(t)
	if err != nil {
		fmt.Println("json Marshal err: ", err)
	}
	fmt.Println(data)

	err = cli.Call(&transcationHash, "eth_sendTransaction", data)
	if err != nil {
		fmt.Println("failed to connect to eth_sendTransaction", err)
		return "", err
	}
	fmt.Println("eth_sendTransaction successfully")
	return transcationHash, err
}

// VoteTo ...
func VoteTo(from, pass string, tokenID int64) error {
	cli, err := ethclient.Dial(configs.Config.Eth.Connstr)
	if err != nil {
		fmt.Println("failed to ethclient.Dial", err)
		return err
	}
	instance, err := NewPxa(common.HexToAddress(configs.Config.Eth.PxaAddr), cli)
	if err != nil {
		fmt.Println("failed to eths.NewPxa", err)
		return err
	}
	// 设置签名, owner的keyStore文件
	// 需要获得文件名字
	fileName, err := utils.GetFileName(string([]rune(from)[2:]), configs.Config.Eth.Keydir)

	file, err := os.Open(configs.Config.Eth.Keydir + "/" + fileName)
	if err != nil {
		fmt.Println("failed to os.Open", err)
		return err
	}
	auth, err := bind.NewTransactor(file, pass)
	if err != nil {
		fmt.Println("failed to bind.NewTransactor", err)
		return err
	}
	// string -> [32]byte
	_, err = instance.Vote(auth, big.NewInt(tokenID))
	if err != nil {
		fmt.Println("failed to Vote", err)
		return err
	}
	StorageVoteCount()

	fmt.Printf("the account: %s Vote success...\n", from)
	return nil
}

// StorageVoteCount ...
func StorageVoteCount() error {
	cli, err := ethclient.Dial(configs.Config.Eth.Connstr)
	if err != nil {
		fmt.Println("failed to ethclient.Dial", err)
		return err
	}
	instance, err := NewPxa(common.HexToAddress(configs.Config.Eth.PxaAddr), cli)
	if err != nil {
		fmt.Println("failed to eths.NewPxa", err)
		return err
	}
	// 查询vote数据库中的token_id 进行遍历

	tokenSQL := fmt.Sprintf("select distinct token_id from vote")
	tokenIds, num, err := dbs.DBQuery(tokenSQL)
	CountStorage = make(map[int64]*big.Int)
	if num > 0 && err == nil {
		// string -> [32]byte
		for _, tokenID := range tokenIds {
			newTokenID, _ := strconv.ParseInt(tokenID["token_id"], 10, 32)
			newAsset, err := instance.Assets(nil, big.NewInt(newTokenID))
			if err != nil {
				fmt.Println("failed to instance.Assets", err)
				return err
			}
			CountStorage[newTokenID] = newAsset.VoteCount
		}
	}
	return err
}

// ViewVoteCount ...
func ViewVoteCount() {
	for k, v := range CountStorage {
		fmt.Printf("token_id: %d ====Count: %d \n", k, v)
	}
}
