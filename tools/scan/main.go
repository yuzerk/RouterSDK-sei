package main

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"

	"github.com/anyswap/CrossChain-Router/v3/common"
	"github.com/anyswap/CrossChain-Router/v3/log"
	"github.com/anyswap/CrossChain-Router/v3/router/bridge"
	"github.com/anyswap/CrossChain-Router/v3/rpc/client"
	"github.com/anyswap/CrossChain-Router/v3/tokens"
	routersdk "github.com/anyswap/RouterSDK-sei/sdk"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	paramMpc  = "cosmos1zc4qe220ceag38j0rwpanjfetp8lpkcvhutksa"
	BlockInfo = "/blocks/"
	urls      = []string{"https://rest.atlantic-2.seinetwork.io:443"}

	br = routersdk.NewCrossChainBridge()
)

func main() {
	br.SetGatewayConfig(&tokens.GatewayConfig{APIAddress: urls})
	testTxHash := "CEA5353DD3FE3BA9880D4B55E29F0AEE0732DE7C086D115B1B2B4FEF7DFA85BE"
	getTransactionDetail(testTxHash)
}

func Scan(blockNum uint64) {
	if res, err := GetBlockByNumber(blockNum); err != nil {
		log.Fatal("GetBlockByNumber error", "err", err)
	} else {
		for _, tx := range res.Block.Data.Txs {
			if txBytes, err := base64.StdEncoding.DecodeString(tx); err == nil {
				txHash := fmt.Sprintf("%X", routersdk.Sha256Sum(txBytes))
				if txRes, err := br.GetTransactionByHash(txHash); err == nil {
					if err := ParseMemo(txRes.Tx.Body.Memo); err == nil {
						if err := ParseAmountTotal(txRes.TxResponse.Logs); err == nil {
							log.Info("verify txHash success", "txHash", txHash)
						}
					}
				}
			}
		}
	}
}

func ParseAmountTotal(messageLogs []sdk.ABCIMessageLog) (err error) {
	for _, logDetail := range messageLogs {
		for _, event := range logDetail.Events {
			if event.Type == routersdk.TransferType {
				if (len(event.Attributes) == 2 || len(event.Attributes) == 3) && event.Attributes[0].Value == paramMpc {
					return nil
				}
			}
		}
	}
	return fmt.Errorf("txHash not match")
}

func ParseMemo(memo string) error {
	fields := strings.Split(memo, ":")
	if len(fields) == 2 {
		if toChainID, err := common.GetBigIntFromStr(fields[1]); err != nil {
			return err
		} else {
			dstBridge := bridge.NewCrossChainBridge(toChainID)
			if dstBridge != nil && dstBridge.IsValidAddress(fields[0]) {
				return nil
			}
		}
	}
	return tokens.ErrTxWithWrongMemo
}

func GetBlockByNumber(blockNumber uint64) (*GetLatestBlockResponse, error) {
	var result *GetLatestBlockResponse
	for _, url := range urls {
		restApi := url + BlockInfo + fmt.Sprint(blockNumber)
		if err := client.RPCGet(&result, restApi); err == nil {
			return result, nil
		}
	}
	return nil, tokens.ErrRPCQueryError
}

func getTransactionDetail(txHash string) {
	if txRes, err := br.GetTransactionByHash(txHash); err == nil {
		if err := ParseMemo(txRes.Tx.Body.Memo); err == nil {
			if err := ParseAmountTotal(txRes.TxResponse.Logs); err == nil {
				log.Info("verify txHash success", "txHash", txHash)
			}
		}
	}
}

func getBalanceOf(address string, result *big.Int) {
	balance, err := br.GetBalance(address)
	if err != nil {
		log.Info("get balance error", "address", address, "error", err)
		*result = *sdk.ZeroInt().BigInt()
	} else {
		log.Info("verify balance amount", "balance", balance)
		*result = *balance
	}
}

func getDenomBalanceOf(address string, denom string, result *big.Int) {
	balance, err := br.GetDenomBalance(address, denom)
	if err != nil {
		log.Info("get denom balance error", "address", address, "denom", denom, "error", err)
		*result = *sdk.ZeroInt().BigInt()
	} else {
		log.Info("verify denom balance amount", " denom balance", balance)
		*result = *balance.BigInt()
	}
}

type GetLatestBlockResponse struct {
	// Deprecated: please use `sdk_block` instead
	Block *Block `protobuf:"bytes,2,opt,name=block,proto3" json:"block,omitempty"`
}

type Block struct {
	Data Data `json:"data"`
}

type Data struct {
	Txs []string `json:"txs"`
}
