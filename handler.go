package main

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"github.com/wolot/gosdk"
	"github.com/wolot/gosdk/types"
	"math/big"
	"strings"
)

type Handler struct {
	cfg                  *Config
	mondoCli             *gosdk.APIClient
	tmpPubKey, tmpPriKey string
	res                  *Resp
}

func NewHandler(cfg *Config) *Handler {
	mondoCli := gosdk.NewAPIClient(cfg.MondoApi)
	pubKey, _, priKey, _ := gosdk.GenKey()
	return &Handler{
		res:       NewResp(),
		cfg:       cfg,
		mondoCli:  mondoCli,
		tmpPriKey: priKey,
		tmpPubKey: pubKey,
	}
}

// @Summary 生成账户
// @Description 生成mondo账户，该账户默认未上链，需要通过转账交易激活
// @Tags v1
// @Accept json
// @Produce json
// @Success 200 {object}  GenKeyResult "成功"
// @Router /v1/genkey [get]
func (h *Handler) GenKey(ctx *gin.Context) {
	pubKey, address, priKey, err := gosdk.GenKey()
	if err != nil {
		h.res.RespError(500, err, ctx)
		return
	}

	data := GenKeyResult{
		PubKey:  pubKey,
		Address: address,
		PriKey:  priKey,
	}
	h.res.RespResult(data, ctx)
}

// @Summary 校验地址是否合法
// @Description 校验地址是否合法
// @Tags v1
// @Accept json
// @Produce json
// @Param address query string true "地址"
// @Success 200 {object} bool "成功"
// @Router /v1/validaddress [get]
func (h *Handler) ValidAddress(ctx *gin.Context) {
	address := ctx.Query("address")
	if address == "" {
		h.res.RespResult(false, ctx)
		return
	}
	h.res.RespResult(gosdk.ValidAddress(address), ctx)
}

// @Summary 查询地址OLO余额
// @Description 查询地址OLO余额，如果地址不存在，返回余额为0
// @Tags v1
// @Accept json
// @Produce json
// @Param address query string true "地址"
// @Success 200 {object} BalanceResult "成功"
// @Router /v1/olobalance [get]
func (h *Handler) OLOBalance(ctx *gin.Context) {
	address := ctx.Query("address")
	if address == "" {
		h.res.Resp(ParamError, "address is nil", ctx)
		return
	}
	bal, err := h.mondoCli.GetBalance(address)
	if err != nil {
		h.res.RespError(SysTemError, err, ctx)
		return
	}
	h.res.RespResult(BalanceResult{Balance: bal.String()}, ctx)
}

// @Summary 查询地址代币余额
// @Description 查询地址代币余额
// @Tags v1
// @Accept json
// @Produce json
// @Param address query string true "地址"
// @Param token query string true "代币合约地址"
// @Success 200 {object} BalanceResult "成功"
// @Router /v1/tokenbalance [get]
func (h *Handler) TokenBalance(ctx *gin.Context) {
	address := ctx.Query("address")
	token := ctx.Query("token")
	if address == "" || token == "" {
		h.res.Resp(ParamError, "address or token is nil", ctx)
		return
	}
	bal, err := h.mondoCli.ERC20BalanceOf(h.tmpPubKey, h.tmpPriKey, token, address)
	if err != nil {
		h.res.RespError(SysTemError, err, ctx)
		return
	}
	h.res.RespResult(BalanceResult{Balance: bal.String()}, ctx)
}

// @Summary 生成OLO转账交易
// @Description 生成OLO转账交易
// @Tags v1
// @Accept json
// @Produce json
// @Param Request body BuildOLOTxReq true "请求参数"
// @Success 200 {object}  BuildTxResult "成功"
// @Router /v1/buildolotx [POST]
func (h *Handler) BuildOLOTx(ctx *gin.Context) {
	var req BuildOLOTxReq
	if err := ctx.BindJSON(&req); err != nil {
		h.res.RespError(SysTemError, err, ctx)
		return
	}
	if err := req.Check(); err != nil {
		h.res.RespError(ParamError, err, ctx)
		return
	}

	privkey, err := crypto.ToECDSA(common.Hex2Bytes(req.PriKey))
	if err != nil {
		h.res.RespError(ParamError, err, ctx)
		return
	}
	publicKey := common.Bytes2Hex(crypto.CompressPubkey(&privkey.PublicKey))

	tx, err := h.mondoCli.BuildEvmTx(publicKey, req.PriKey, req.To, req.Value, "", 210000, "1", "")
	if err != nil {
		h.res.RespError(SysTemError, err, ctx)
		return
	}
	h.res.RespResult(BuildTxResult{
		Hash:        tx.Hash().Hex(),
		SignedEvmTx: *txToJson(tx),
	}, ctx)
}

// @Summary 生成代币转账交易
// @Description 生成代币转账交易
// @Tags v1
// @Accept json
// @Produce json
// @Param Request body BuildTokenTxReq true "请求参数"
// @Success 200 {object}  BuildTxResult "成功"
// @Router /v1/buildtokentx [POST]
func (h *Handler) BuildTokenTx(ctx *gin.Context) {
	var req BuildTokenTxReq
	if err := ctx.BindJSON(&req); err != nil {
		h.res.RespError(SysTemError, err, ctx)
		return
	}
	if err := req.Check(); err != nil {
		h.res.RespError(ParamError, err, ctx)
		return
	}

	privkey, err := crypto.ToECDSA(common.Hex2Bytes(req.PriKey))
	if err != nil {
		h.res.RespError(ParamError, err, ctx)
		return
	}
	publicKey := common.Bytes2Hex(crypto.CompressPubkey(&privkey.PublicKey))

	abiIns, _ := abi.JSON(strings.NewReader(ABIJSON))
	value, _ := new(big.Int).SetString(req.Value, 10)
	bz, _ := abiIns.Pack("transfer", common.HexToAddress(req.To), value)

	tx, err := h.mondoCli.BuildEvmTx(publicKey, req.PriKey, req.Token, "0", hex.EncodeToString(bz), 100000000, "1", "")
	if err != nil {
		h.res.RespError(SysTemError, err, ctx)
		return
	}
	h.res.RespResult(BuildTxResult{
		Hash:        tx.Hash().Hex(),
		SignedEvmTx: *txToJson(tx),
	}, ctx)
}

// @Summary 生成增发代币交易
// @Description 生成增发代币交易
// @Tags v1
// @Accept json
// @Produce json
// @Param Request body BuildTokenIssueTxReq true "请求参数"
// @Success 200 {object}  BuildTxResult "成功"
// @Router /v1/buildtokenissuetx [POST]
func (h *Handler) BuildTokenIssueTx(ctx *gin.Context) {
	var req BuildTokenIssueTxReq
	if err := ctx.BindJSON(&req); err != nil {
		h.res.RespError(SysTemError, err, ctx)
		return
	}
	if err := req.Check(); err != nil {
		h.res.RespError(ParamError, err, ctx)
		return
	}

	privkey, err := crypto.ToECDSA(common.Hex2Bytes(req.PriKey))
	if err != nil {
		h.res.RespError(ParamError, err, ctx)
		return
	}
	publicKey := common.Bytes2Hex(crypto.CompressPubkey(&privkey.PublicKey))

	abiIns, _ := abi.JSON(strings.NewReader(ABIJSON))
	value, _ := new(big.Int).SetString(req.Value, 10)
	bz, _ := abiIns.Pack("issue", value)

	tx, err := h.mondoCli.BuildEvmTx(publicKey, req.PriKey, req.Token, "0", hex.EncodeToString(bz), 100000000, "1", "")
	if err != nil {
		h.res.RespError(SysTemError, err, ctx)
		return
	}
	h.res.RespResult(BuildTxResult{
		Hash:        tx.Hash().Hex(),
		SignedEvmTx: *txToJson(tx),
	}, ctx)
}

// @Summary 生成增发代币交易
// @Description 生成增发代币交易
// @Tags v1
// @Accept json
// @Produce json
// @Param Request body BuildTokenIssueTxReq true "请求参数"
// @Success 200 {object}  BuildTxResult "成功"
// @Router /v1/buildtokenredeemtx [POST]
func (h *Handler) BuildTokenRedeemTx(ctx *gin.Context) {
	var req BuildTokenIssueTxReq
	if err := ctx.BindJSON(&req); err != nil {
		h.res.RespError(SysTemError, err, ctx)
		return
	}
	if err := req.Check(); err != nil {
		h.res.RespError(ParamError, err, ctx)
		return
	}

	privkey, err := crypto.ToECDSA(common.Hex2Bytes(req.PriKey))
	if err != nil {
		h.res.RespError(ParamError, err, ctx)
		return
	}
	publicKey := common.Bytes2Hex(crypto.CompressPubkey(&privkey.PublicKey))

	abiIns, _ := abi.JSON(strings.NewReader(ABIJSON))
	value, _ := new(big.Int).SetString(req.Value, 10)
	bz, _ := abiIns.Pack("redeem", value)

	tx, err := h.mondoCli.BuildEvmTx(publicKey, req.PriKey, req.Token, "0", hex.EncodeToString(bz), 100000000, "1", "")
	if err != nil {
		h.res.RespError(SysTemError, err, ctx)
		return
	}
	h.res.RespResult(BuildTxResult{
		Hash:        tx.Hash().Hex(),
		SignedEvmTx: *txToJson(tx),
	}, ctx)
}

// @Summary 生成代币批量转账交易
// @Description 生成代币批量转账交易，转账金额不同
// @Tags v1
// @Accept json
// @Produce json
// @Param Request body BuildTokenBatchTransfersTxReq true "请求参数"
// @Success 200 {object}  BuildTxResult "成功"
// @Router /v1/buildtokenbatchtxs [POST]
func (h *Handler) BuildTokenBatchTransfersTx(ctx *gin.Context) {
	var req BuildTokenBatchTransfersTxReq
	if err := ctx.BindJSON(&req); err != nil {
		h.res.RespError(SysTemError, err, ctx)
		return
	}
	if err := req.Check(); err != nil {
		h.res.RespError(ParamError, err, ctx)
		return
	}

	privkey, err := crypto.ToECDSA(common.Hex2Bytes(req.PriKey))
	if err != nil {
		h.res.RespError(ParamError, err, ctx)
		return
	}
	publicKey := common.Bytes2Hex(crypto.CompressPubkey(&privkey.PublicKey))

	var tos []common.Address
	var values []*big.Int
	var gasLimit = uint64(70000 * len(req.Tos))
	for _, v := range req.Tos {
		tos = append(tos, common.HexToAddress(v.To))
		value, _ := new(big.Int).SetString(v.Value, 10)
		values = append(values, value)
	}
	abiIns, _ := abi.JSON(strings.NewReader(ABIJSON))
	bz, _ := abiIns.Pack("batchTransfers", tos, values)

	tx, err := h.mondoCli.BuildEvmTx(publicKey, req.PriKey, req.Token, "0", hex.EncodeToString(bz), gasLimit, "1", "")
	if err != nil {
		h.res.RespError(SysTemError, err, ctx)
		return
	}
	h.res.RespResult(BuildTxResult{
		Hash:        tx.Hash().Hex(),
		SignedEvmTx: *txToJson(tx),
	}, ctx)
}

// @Summary 生成代币批量转账交易
// @Description 生成代币批量转账交易，转账金额相同
// @Tags v1
// @Accept json
// @Produce json
// @Param Request body BuildTokenBatchTransferTxReq true "请求参数"
// @Success 200 {object}  BuildTxResult "成功"
// @Router /v1/buildtokenbatchtx [POST]
func (h *Handler) BuildTokenBatchTransferTx(ctx *gin.Context) {
	var req BuildTokenBatchTransferTxReq
	if err := ctx.BindJSON(&req); err != nil {
		h.res.RespError(SysTemError, err, ctx)
		return
	}
	if err := req.Check(); err != nil {
		h.res.RespError(ParamError, err, ctx)
		return
	}

	privkey, err := crypto.ToECDSA(common.Hex2Bytes(req.PriKey))
	if err != nil {
		h.res.RespError(ParamError, err, ctx)
		return
	}
	publicKey := common.Bytes2Hex(crypto.CompressPubkey(&privkey.PublicKey))

	var tos []common.Address
	var gasLimit = uint64(70000 * len(req.Tos))
	for _, v := range req.Tos {
		tos = append(tos, common.HexToAddress(v))

	}
	abiIns, _ := abi.JSON(strings.NewReader(ABIJSON))
	value, _ := new(big.Int).SetString(req.Value, 10)
	bz, _ := abiIns.Pack("batchTransfer", tos, value)

	tx, err := h.mondoCli.BuildEvmTx(publicKey, req.PriKey, req.Token, "0", hex.EncodeToString(bz), gasLimit, "1", "")
	if err != nil {
		h.res.RespError(SysTemError, err, ctx)
		return
	}
	h.res.RespResult(BuildTxResult{
		Hash:        tx.Hash().Hex(),
		SignedEvmTx: *txToJson(tx),
	}, ctx)
}

// @Summary 发送交易
// @Description 发送交易
// @Tags v1
// @Accept json
// @Produce json
// @Param Request body SignedEvmTx true "请求参数"
// @Success 200 "成功"
// @Router /v1/sendtx [POST]
func (h *Handler) SendTx(ctx *gin.Context) {
	var req gosdk.SignedEvmTx
	if err := ctx.BindJSON(&req); err != nil {
		h.res.RespError(SysTemError, err, ctx)
		return
	}
	if err := h.mondoCli.SendEvmTx("0", jsonToTx(&req)); err != nil {
		h.res.RespError(SysTemError, err, ctx)
		return
	}
	h.res.RespOk(ctx)
}

// @Summary 检测交易是否成功上链
// @Description 检测交易是否成功上链
// @Tags v1
// @Accept json
// @Produce json
// @Param hash query string true "交易hash"
// @Success 200 {object} bool "成功"
// @Router /v1/checktx [get]
func (h *Handler) CheckTx(ctx *gin.Context) {
	hash := ctx.Query("hash")
	tx, err := h.mondoCli.V3GetTransaction(hash)
	if err != nil {
		h.res.RespError(SysTemError, err, ctx)
		return
	}
	if tx != nil && tx.Codei == 0 {
		h.res.RespResult(true, ctx)
		return
	}
	h.res.RespResult(false, ctx)
}

func txToJson(tx *types.TxEvm) *gosdk.SignedEvmTx {
	var jsonTx gosdk.SignedEvmTx
	jsonTx.CreatedAt = tx.CreatedAt
	jsonTx.GasLimit = tx.GasLimit
	jsonTx.GasPrice = tx.GasPrice.String()
	jsonTx.Nonce = tx.Nonce
	jsonTx.Sender = tx.Sender.String()
	jsonTx.Body.To = tx.Body.To.ToAddress().String()
	jsonTx.Body.Value = tx.Body.Value.String()
	jsonTx.Body.Load = hex.EncodeToString(tx.Body.Load)
	jsonTx.Body.Memo = string(tx.Body.Memo)
	jsonTx.Signature = hex.EncodeToString(tx.Signature)
	return &jsonTx
}

func jsonToTx(json *gosdk.SignedEvmTx) *types.TxEvm {
	tx := types.NewTxEvm()
	tx.CreatedAt = json.CreatedAt
	tx.GasLimit = json.GasLimit
	tx.GasPrice, _ = new(big.Int).SetString(json.GasPrice, 10)
	tx.Nonce = json.Nonce
	tx.Sender, _ = types.HexToPubkey(json.Sender)
	tx.Body.To.SetBytes(common.HexToAddress(json.Body.To).Bytes())
	tx.Body.Value, _ = new(big.Int).SetString(json.Body.Value, 10)
	tx.Body.Load = HexToBytes(json.Body.Load)
	tx.Body.Memo = []byte(json.Body.Memo)
	tx.Signature = HexToBytes(json.Signature)
	return tx
}

func HexToBytes(str string) []byte {
	str = strings.TrimPrefix(str, "0x")
	b, _ := hex.DecodeString(str)
	return b
}
