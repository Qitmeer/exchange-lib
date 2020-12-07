package rpc

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type Client struct {
	rpcCfg *RpcConfig
}

type RpcConfig struct {
	Address string
	User    string
	Pwd     string
	Https   bool
}

func NewClient(cfg *RpcConfig) *Client {
	return &Client{
		cfg,
	}
}

func (c *Client) GetBlockByOrder(order uint64) (*Block, error) {
	params := []interface{}{order, true}
	resp, err := NewReqeust(params).SetMethod("getBlockByOrder").call(c.rpcCfg)
	if err != nil {
		return nil, err
	}
	blk := new(Block)
	if resp.Error != nil {
		return blk, errors.New(resp.Error.Message)
	}
	if err := json.Unmarshal(resp.Result, blk); err != nil {
		return blk, errors.New("failed to parse response json")
	}
	return blk, nil
}

func (c *Client) GetBlockCount() string {
	var params []interface{}
	resp, err := NewReqeust(params).SetMethod("getBlockCount").call(c.rpcCfg)
	if err != nil {
		return "-1"
	}
	if resp.Error != nil {
		return "-1"
	}
	return string(resp.Result)
}

func (c *Client) SendTransaction(tx string) (string, error) {
	params := []interface{}{strings.Trim(tx, "\n"), false}
	resp, err := NewReqeust(params).SetMethod("sendRawTransaction").call(c.rpcCfg)
	if err != nil {
		return "", err
	}
	if resp.Error != nil {
		return resp.Error.Message, errors.New(resp.Error.Message)
	}
	txid := ""
	json.Unmarshal(resp.Result, &txid)
	return txid, nil
}

func (c *Client) GetTransaction(txId string) (*Transaction, error) {
	params := []interface{}{txId, true}
	resp, err := NewReqeust(params).SetMethod("getRawTransaction").call(c.rpcCfg)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, errors.New(resp.Error.Message)
	}
	var rs *Transaction
	if err := json.Unmarshal(resp.Result, &rs); err != nil {
		return nil, errors.New("failed to parse response json")
	}
	return rs, nil
}

func (c *Client) GetMemoryPool() ([]string, error) {
	params := []interface{}{"", false}
	resp, err := NewReqeust(params).SetMethod("getMempool").call(c.rpcCfg)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, errors.New(resp.Error.Message)
	}
	var rs []string
	if err := json.Unmarshal(resp.Result, &rs); err != nil {
		return nil, err
	}
	return rs, nil
}

func (c *Client) GetBlockById(id uint64) (*Block, error) {
	params := []interface{}{id, true}
	resp, err := NewReqeust(params).SetMethod("getBlockByID").call(c.rpcCfg)
	if err != nil {
		return nil, err
	}
	blk := new(Block)
	if resp.Error != nil {
		return blk, errors.New(resp.Error.Message)
	}
	if err := json.Unmarshal(resp.Result, blk); err != nil {
		return blk, errors.New("failed to parse response json")
	}
	return blk, nil
}

func (c *Client) GetNodeInfo() (*NodeInfo, error) {
	params := []interface{}{}
	resp, err := NewReqeust(params).SetMethod("getNodeInfo").call(c.rpcCfg)
	if err != nil {
		return nil, err
	}
	nodeInfo := new(NodeInfo)
	if resp.Error != nil {
		return nodeInfo, errors.New(resp.Error.Message)
	}
	if err := json.Unmarshal(resp.Result, nodeInfo); err != nil {
		return nodeInfo, errors.New("failed to parse response json")
	}
	return nodeInfo, nil
}

func (c *Client) IsBlue(hash string) (int, error) {
	params := []interface{}{hash}
	resp, err := NewReqeust(params).SetMethod("isBlue").call(c.rpcCfg)
	if err != nil {
		return 0, err
	}
	if resp.Error != nil {
		return 0, errors.New(resp.Error.Message)
	}
	state, err := strconv.Atoi(string(resp.Result))
	if err != nil {
		return 0, err
	}
	return state, nil
}

func (req *ClientRequest) call(rpcCfg *RpcConfig) (*ClientResponse, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	defer client.CloseIdleConnections()

	//convert struct to []byte
	marshaledData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("rpc client encoding json failed; error:%s ", err.Error())
	}
	httpUrl := "http://"
	if rpcCfg.Https {
		httpUrl = "https://"
	}

	httpRequest, err :=
		http.NewRequest(http.MethodPost, httpUrl+rpcCfg.Address, bytes.NewReader(marshaledData))
	if err != nil {
		return nil, fmt.Errorf("rpc client create request failed; error:%s ", err.Error())
	}
	if httpRequest == nil {
		return nil, fmt.Errorf("rpc client create request failed")
	}
	httpRequest.Close = true
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.SetBasicAuth(rpcCfg.User, rpcCfg.Pwd)

	response, err := client.Do(httpRequest)
	if err != nil {
		return &ClientResponse{Error: &Error{Message: err.Error()}}, nil
	}

	body := response.Body

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body; error:%s", err.Error())
	}

	resp := &ClientResponse{}
	//convert []byte to struct
	if err := json.Unmarshal(bodyBytes, resp); err != nil {
		return nil, fmt.Errorf("json unmarshal failed; value:%s; error:%s", string(bodyBytes), err.Error())
	}

	err = response.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("close response failed; error:%s", err.Error())
	}

	return resp, nil
}
