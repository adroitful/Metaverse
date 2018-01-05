package rpc

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"

	"github.com/sammy007/open-ethereum-pool/util"
)

type GetBalanceReply struct {
	Unspent string `json:"unspent"`
}

type SendTransactionReply struct {
	Hash string `json:"hash"`
}

type RPCClient struct {
	sync.RWMutex
	Url         string
	Name        string
	Account     string
	Password    string
	sick        bool
	sickRate    int
	successRate int
	client      *http.Client
}

/*
type GetBlockReply struct {
	Number string `json:"number"`
	Hash   string `json:"hash"`
	Nonce  string `json:"nonce"`
	//Miner        string   `json:"miner"`
	Difficulty   string   `json:"difficulty"`
	GasLimit     string   `json:"gasLimit"`
	GasUsed      string   `json:"gasUsed"`
	Transactions []Tx     `json:"transactions"`
	Uncles       []string `json:"uncles"`
	// https://github.com/ethereum/EIPs/issues/95
	SealFields []string `json:"sealFields"`
}
*/
type GetBlockReply struct {
	Difficulty       string `json:"bits"`
	Hash             string `json:"hash"`
	MerkleTreeHash   string `json:"merkle_tree_hash"`
	Nonce            string `json:"nonce"`
	PrevHash         string `json:"previous_block_hash"`
	TimeStamp        string `json:"time_stamp"`
	Version          string `json:"version"`
	Mixhash          string `json:"mixhash"`
	Number           string `json:"number"`
	TransactionCount string `json:"transaction_count"`
}

type GetBlockReplyPart struct {
	Number     string `json:"number"`
	Difficulty string `json:"bits"`
}

type TxReceipt struct {
	TxHash string `json:"hash"`
	//GasUsed   string `json:"gasUsed"`
	//BlockHash string `json:"blockHash"`
}

/*
func (r *TxReceipt) Confirmed() bool {
	return len(r.BlockHash) > 0
}
*/

type Tx struct {
	Gas      string `json:"gas"`
	GasPrice string `json:"gasPrice"`
	Hash     string `json:"hash"`
}

type JSONRpcResp struct {
	Id          *json.RawMessage       `json:"id"`
	Result      *json.RawMessage       `json:"result"`
	Balance     *json.RawMessage       `json:"balance"`
	Transaction *json.RawMessage       `json:"transaction"`
	Error       map[string]interface{} `json:"error"`
}

func NewRPCClient(name, url, account, password, timeout string) *RPCClient {
	rpcClient := &RPCClient{Name: name, Url: url, Account: account, Password: password}
	timeoutIntv := util.MustParseDuration(timeout)
	rpcClient.client = &http.Client{
		Timeout: timeoutIntv,
	}
	return rpcClient
}

func (r *RPCClient) GetWork() ([]string, error) {
	//rpcResp, err := r.doPost(r.Url, "eth_getWork", []string{})
	rpcResp, err := r.doPost(r.Url, "getwork", []string{})
	if err != nil {
		return nil, err
	}
	var reply []string
	err = json.Unmarshal(*rpcResp.Result, &reply)
	return reply, err
}

func (r *RPCClient) SetAddress(address string) ([]string, error) {
	rpcResp, err := r.doPost(r.Url, "setminingaccount", []string{r.Account, r.Password, address})
	if err != nil {
		return nil, err
	}
	var reply []string
	err = json.Unmarshal(*rpcResp.Result, &reply)
	return reply, err
}

func (r *RPCClient) GetHeight() (int64, error) {
	rpcResp, err := r.doPost(r.Url, "fetch-height", []string{})
	if err != nil {
		return 0, err
	}
	var height int64
	err = json.Unmarshal(*rpcResp.Result, &height)
	return height, err
}

func (r *RPCClient) GetPendingBlock() (*GetBlockReplyPart, error) {
	rpcResp, err := r.doPost(r.Url, "fetchheaderext", []interface{}{r.Account, r.Password, "pending"})
	if err != nil {
		return nil, err
	}
	if rpcResp.Result != nil {
		var reply *GetBlockReplyPart
		err = json.Unmarshal(*rpcResp.Result, &reply)
		return reply, err
	}
	return nil, nil
}

func (r *RPCClient) GetBlockByHeight(height int64) (*GetBlockReply, error) {
	//params := []interface{}{fmt.Sprintf("0x%x", height), true}
	params := []interface{}{"-t", height}
	return r.getBlockBy("fetch-header", params)
}

func (r *RPCClient) GetBlockByHash(hash string) (*GetBlockReply, error) {
	params := []interface{}{hash, true}
	return r.getBlockBy("eth_getBlockByHash", params)
}

func (r *RPCClient) GetUncleByBlockNumberAndIndex(height int64, index int) (*GetBlockReply, error) {
	params := []interface{}{fmt.Sprintf("0x%x", height), fmt.Sprintf("0x%x", index)}
	return r.getBlockBy("eth_getUncleByBlockNumberAndIndex", params)
}

func (r *RPCClient) getBlockBy(method string, params []interface{}) (*GetBlockReply, error) {
	rpcResp, err := r.doPost(r.Url, method, params)
	if err != nil {
		return nil, err
	}
	if rpcResp.Result != nil {
		var reply *GetBlockReply
		err = json.Unmarshal(*rpcResp.Result, &reply)
		return reply, err
	}
	return nil, nil
}

func (r *RPCClient) GetTxReceipt(hash string) (*TxReceipt, error) {
	rpcResp, err := r.doPost(r.Url, "fetch-tx", []string{hash})
	if err != nil {
		return nil, err
	}
	if rpcResp.Result != nil {
		var reply *TxReceipt
		err = json.Unmarshal(*rpcResp.Transaction, &reply)
		return reply, err
	}
	return nil, nil
}

func (r *RPCClient) validateaddress(hash string) (*TxReceipt, error) {
	rpcResp, err := r.doPost(r.Url, "validateaddress", params)
	if err != nil {
		return reply, err
	}
	if rpcResp.Result != nil {
		var reply *ValidAddress
		err = json.Unmarshal(*rpcResp.Transaction, &reply)
		return reply, err
	}
	return nil, nil
}

func (r *RPCClient) SubmitBlock(params []string) (bool, error) {
	//rpcResp, err := r.doPost(r.Url, "eth_submitWork", params)
	rpcResp, err := r.doPost(r.Url, "submitwork", params)
	if err != nil {
		return false, err
	}
	var reply bool
	//err = json.Unmarshal(*rpcResp.Result, &reply_str)
	fmt.Println(*rpcResp.Result)
	if string(*rpcResp.Result) == "\"false\"" {
		reply = false
	} else {
		reply = true
	}
	return reply, err
}

func (r *RPCClient) GetBalance(address string) (*big.Int, error) {
	rpcResp, err := r.doPost(r.Url, "fetch-balance", []string{address})
	if err != nil {
		return nil, err
	}
	var reply GetBalanceReply
	err = json.Unmarshal(*rpcResp.Balance, &reply)
	if err != nil {
		return nil, err
	}
	return common.String2Big(reply.Unspent), err
}

func (r *RPCClient) Sign(from string, s string) (string, error) {
	hash := sha256.Sum256([]byte(s))
	rpcResp, err := r.doPost(r.Url, "eth_sign", []string{from, common.ToHex(hash[:])})
	var reply string
	if err != nil {
		return reply, err
	}
	err = json.Unmarshal(*rpcResp.Result, &reply)
	if err != nil {
		return reply, err
	}
	if util.IsZeroHash(reply) {
		err = errors.New("Can't sign message, perhaps account is locked")
	}
	return reply, err
}

func (r *RPCClient) GetPeerCount() (int64, error) {
	rpcResp, err := r.doPost(r.Url, "net_peerCount", nil)
	if err != nil {
		return 0, err
	}
	var reply string
	err = json.Unmarshal(*rpcResp.Result, &reply)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.Replace(reply, "0x", "", -1), 16, 64)
}

func (r *RPCClient) validateaddress(address string) (*big.Int, error) {
	rpcResp, err := r.doPost(r.Url, "validateaddress", []string{address})
	if err != nil {
		return reply, err
	}
	var reply ValidateReply
	err = json.Unmarshal(*rpcResp.ValidateReply, &reply)
	if err != nil {
		return nil, err
	}
	return reply, err

func (r *RPCClient) SendTransaction(from, to, value string) (string, error) {
	rpcResp, err := r.doPost(r.Url, "sendfrom", []string{r.Account, r.Password, from, to, value})
	var reply SendTransactionReply
	if err != nil {
		return reply.Hash, err
	}
	err = json.Unmarshal(*rpcResp.Transaction, &reply)
	if err != nil {
		return reply.Hash, err
	}
	/* There is an inconsistence in a "standard". Geth returns error if it can't unlock signer account,
	 * but Parity returns zero hash 0x000... if it can't send tx, so we must handle this case.
	 * https://github.com/ethereum/wiki/wiki/JSON-RPC#returns-22
	 */
	/*
		if util.IsZeroHash(reply) {
			err = errors.New("transaction is not yet available")
		}
	*/
	return reply.Hash, err
}

func (r *RPCClient) doPost(url string, method string, params interface{}) (*JSONRpcResp, error) {
	jsonReq := map[string]interface{}{"jsonrpc": "2.0", "method": method, "params": params, "id": 0}
	data, _ := json.Marshal(jsonReq)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Content-Length", (string)(len(data)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		r.markSick()
		return nil, err
	}
	defer resp.Body.Close()

	var rpcResp *JSONRpcResp
	err = json.NewDecoder(resp.Body).Decode(&rpcResp)
	if err != nil {
		r.markSick()
		return nil, err
	}
	if rpcResp.Error != nil {
		r.markSick()
		return nil, errors.New(rpcResp.Error["message"].(string))
	}
	return rpcResp, err
}

func (r *RPCClient) Check() bool {
	_, err := r.GetWork()
	if err != nil {
		return false
	}
	r.markAlive()
	return !r.Sick()
}

func (r *RPCClient) Sick() bool {
	r.RLock()
	defer r.RUnlock()
	return r.sick
}

func (r *RPCClient) markSick() {
	r.Lock()
	r.sickRate++
	r.successRate = 0
	if r.sickRate >= 5 {
		r.sick = true
	}
	r.Unlock()
}

func (r *RPCClient) markAlive() {
	r.Lock()
	r.successRate++
	if r.successRate >= 5 {
		r.sick = false
		r.sickRate = 0
		r.successRate = 0
	}
	r.Unlock()
}
