package core

import (
	"github.com/MinterTeam/explorer-gate/env"
	"github.com/MinterTeam/explorer-gate/errors"
	"github.com/daniildulin/minter-node-api"
	"github.com/olebedev/emitter"
	"github.com/sirupsen/logrus"
	"strings"
)

type MinterGate struct {
	api     *minter_node_api.MinterNodeApi
	Config  env.Config
	emitter *emitter.Emitter
	Logger  *logrus.Entry
}

type CoinEstimate struct {
	Value      string
	Commission string
}

//New instance of Minter Gate
func New(config env.Config, e *emitter.Emitter, logger *logrus.Entry) *MinterGate {

	proto := `http`
	if config.GetBool(`minterApi.isSecure`) {
		proto = `https`
	}
	apiLink := proto + `://` + config.GetString(`minterApi.link`) + `:` + config.GetString(`minterApi.port`)
	return &MinterGate{
		emitter: e,
		api:     minter_node_api.New(apiLink),
		Config:  config,
		Logger:  logger,
	}
}

//Send transaction to blockchain
//Return transaction hash
func (mg MinterGate) TxPush(transaction string) (*string, error) {
	response, err := mg.api.PushTransaction(transaction)
	if err != nil {
		mg.Logger.WithFields(logrus.Fields{
			"transaction": transaction,
		}).Error(err)
		return nil, err
	}
	if response.Error != nil || response.Result.Code != 0 {
		err = errors.GetNodeErrorFromResponse(response)
		mg.Logger.WithFields(logrus.Fields{
			"transaction": transaction,
		}).Error(err)
		return nil, err
	}
	hash := `Mt` + strings.ToLower(response.Result.Hash)
	return &hash, nil
}

//Return estimate of transaction
func (mg *MinterGate) EstimateTxCommission(transaction string) (*string, error) {
	response, err := mg.api.GetEstimateTx(transaction)
	if err != nil {
		mg.Logger.WithFields(logrus.Fields{
			"transaction": transaction,
		}).Error(err)
		return nil, err
	}
	if response.Error != nil {
		err = errors.NewNodeError(response.Error.Message, response.Error.Code)
		mg.Logger.WithFields(logrus.Fields{
			"transaction": transaction,
		}).Error(err)
		return nil, err
	}
	return &response.Result.Commission, nil
}

//Return estimate of buy coin
func (mg *MinterGate) EstimateCoinBuy(coinToSell string, coinToBuy string, value string) (*CoinEstimate, error) {
	response, err := mg.api.GetEstimateCoinBuy(coinToSell, coinToBuy, value)
	if err != nil {
		mg.Logger.WithFields(logrus.Fields{
			"coinToSell": coinToSell,
			"coinToBuy":  coinToBuy,
			"value":      value,
		}).Error(err)
		return nil, err
	}

	if response.Error != nil {
		err = errors.NewNodeError(response.Error.Message, response.Error.Code)
		mg.Logger.WithFields(logrus.Fields{
			"coinToSell": coinToSell,
			"coinToBuy":  coinToBuy,
			"value":      value,
		}).Error(err)
		return nil, err
	}

	return &CoinEstimate{response.Result.WillPay, response.Result.Commission}, nil
}

//Return estimate of sell coin
func (mg *MinterGate) EstimateCoinSell(coinToSell string, coinToBuy string, value string) (*CoinEstimate, error) {
	response, err := mg.api.GetEstimateCoinSell(coinToSell, coinToBuy, value)
	if err != nil {
		mg.Logger.WithFields(logrus.Fields{
			"coinToSell": coinToSell,
			"coinToBuy":  coinToBuy,
			"value":      value,
		}).Error(err)
		return nil, err
	}
	if response.Error != nil {
		err = errors.NewNodeError(response.Error.Message, response.Error.Code)
		mg.Logger.WithFields(logrus.Fields{
			"coinToSell": coinToSell,
			"coinToBuy":  coinToBuy,
			"value":      value,
		}).Error(err)
		return nil, err
	}
	return &CoinEstimate{response.Result.WillGet, response.Result.Commission}, nil
}

//Return nonce for address
func (mg *MinterGate) GetNonce(address string) (*string, error) {
	response, err := mg.api.GetAddress(address)
	if err != nil {
		mg.Logger.WithFields(logrus.Fields{
			"address": address,
		}).Error(err)
		return nil, err
	}
	if response.Error != nil {
		err = errors.NewNodeError(response.Error.Message, response.Error.Code)
		mg.Logger.WithFields(logrus.Fields{
			"address": address,
		}).Error(err)
		return nil, err
	}
	return &response.Result.TransactionCount, nil
}

//Return nonce for address
func (mg *MinterGate) GetMinGas() (*string, error) {
	response, err := mg.api.GetMinGasPrice()
	if err != nil {
		mg.Logger.Error(err)
		return nil, err
	}
	if response.Error != nil {
		err = errors.NewNodeError(response.Error.Message, response.Error.Code)
		mg.Logger.Error(err)
		return nil, err
	}
	return &response.Result, nil
}
