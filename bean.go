package main

import (
	"errors"
	"github.com/wolot/gosdk"
	"math/big"
)

type GenKeyResult struct {
	PubKey  string `json:"pubKey"`
	Address string `json:"address"`
	PriKey  string `json:"priKey"`
}

type BalanceResult struct {
	Balance string `json:"balance"`
}

type BuildOLOTxReq struct {
	PriKey string `json:"priKey"`
	To     string `json:"to"`
	Value  string `json:"value"`
}

func (p *BuildOLOTxReq) Check() error {
	if p.PriKey == "" || p.To == "" || p.Value == "" {
		return errors.New("got nil param")
	}
	_, b := new(big.Int).SetString(p.Value, 10)
	if !b {
		return errors.New("value error")
	}
	return nil
}

type BuildTokenTxReq struct {
	PriKey string `json:"priKey"`
	Token  string `json:"token"`
	To     string `json:"to"`
	Value  string `json:"value"`
}

func (p *BuildTokenTxReq) Check() error {
	if p.PriKey == "" || p.To == "" || p.Token == "" || p.Value == "" {
		return errors.New("got nil param")
	}
	_, b := new(big.Int).SetString(p.Value, 10)
	if !b {
		return errors.New("value error")
	}
	return nil
}

type BuildTokenIssueTxReq struct {
	PriKey string `json:"priKey"`
	Token  string `json:"token"`
	Value  string `json:"value"`
}

func (p *BuildTokenIssueTxReq) Check() error {
	if p.PriKey == "" || p.Token == "" || p.Value == "" {
		return errors.New("got nil param")
	}
	_, b := new(big.Int).SetString(p.Value, 10)
	if !b {
		return errors.New("value error")
	}
	return nil
}

type BuildTokenBatchTransfersTxReq struct {
	PriKey string `json:"priKey"`
	Token  string `json:"token"`
	Tos    []To   `json:"tos"`
}

func (p *BuildTokenBatchTransfersTxReq) Check() error {
	if p.PriKey == "" || p.Token == "" {
		return errors.New("got nil param")
	}
	for _, v := range p.Tos {
		if v.To == "" {
			return errors.New("got nil to")
		}
		_, b := new(big.Int).SetString(v.Value, 10)
		if !b {
			return errors.New("value error")
		}
	}

	return nil
}

type To struct {
	To    string `json:"to"`
	Value string `json:"value"`
}

type BuildTokenBatchTransferTxReq struct {
	PriKey string   `json:"priKey"`
	Token  string   `json:"token"`
	Tos    []string `json:"tos"`
	Value  string   `json:"value"`
}

func (p *BuildTokenBatchTransferTxReq) Check() error {
	if p.PriKey == "" || p.Token == "" {
		return errors.New("got nil param")
	}
	for _, v := range p.Tos {
		if v == "" {
			return errors.New("got nil to")
		}
	}
	_, b := new(big.Int).SetString(p.Value, 10)
	if !b {
		return errors.New("value error")
	}
	return nil
}

type BuildTxResult struct {
	Hash        string            `json:"hash"`
	SignedEvmTx gosdk.SignedEvmTx `json:"signedEvmTx"`
}

type SignedEvmTx struct {
	Mode      int    `json:"mode"`      // 模式:0-default/commit 1-async 2-sync
	CreatedAt uint64 `json:"createdAt"` // 时间戳unixNano
	GasLimit  uint64 `json:"gasLimit"`  //
	GasPrice  string `json:"gasPrice"`  //
	Nonce     uint64 `json:"nonce"`     //
	Sender    string `json:"sender"`    // pubkey
	Body      struct {
		To    string `json:"to"`    // 合约地址
		Value string `json:"value"` //
		Load  string `json:"load"`  // hex编码
		Memo  string `json:"memo"`  // 备注
	} `json:"body"`
	Signature string `json:"signature"` // hex编码
}
