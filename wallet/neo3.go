/*
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The  poly network  is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The  poly network  is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 * You should have received a copy of the GNU Lesser General Public License
 * along with The poly network .  If not, see <http://www.gnu.org/licenses/>.
 */

package wallet

import (
	"fmt"
	"github.com/joeqian10/neo3-gogogo/crypto"
	"github.com/joeqian10/neo3-gogogo/tx"

	"github.com/joeqian10/neo3-gogogo/helper"
	"github.com/joeqian10/neo3-gogogo/wallet"
	"github.com/polynetwork/bridge-common/base"

	"github.com/polynetwork/bridge-common/chains/neo3"

	"github.com/polynetwork/bridge-common/log"
)

type Neo3Wallet struct {
	sdk *neo3.SDK
	*wallet.NEP6Wallet
	config *Config
}

func NewNeo3Wallet(config *Config, sdk *neo3.SDK) *Neo3Wallet {
	ps := helper.ProtocolSettings{
		Magic:          base.Neo3MagicNum,
		AddressVersion: helper.DefaultAddressVersion,
	}
	w, err := wallet.NewNEP6Wallet(config.Path, &ps, nil, nil)

	if err != nil {
		log.Error("Failed to load neo3 wallet file", "err", err)
		return nil
	}
	err = w.Unlock(config.Password)
	if err != nil {
		log.Error("Failed to decrypt neo3 wallet with password", "err", err)
		return nil
	}
	return &Neo3Wallet{sdk: sdk, NEP6Wallet: w, config: config}
}

func (w *Neo3Wallet) Invoke(script []byte) (hash string, err error) {
	return w.SendInvocation(script)
}

func (w *Neo3Wallet) InvokeWithAccount(nep6Wallet *wallet.NEP6Wallet, script []byte) (hash string, err error) {
	return w.SendInvocationWithAccount(nep6Wallet, script)
}

func (w *Neo3Wallet) SendInvocation(script []byte) (hash string, err error) {
	return w.SendInvocationWithAccount(w.NEP6Wallet, script)
}

func (w *Neo3Wallet) SendInvocationWithAccount(
	nep6Wallet *wallet.NEP6Wallet, script []byte) (hash string, err error) {
	client := w.sdk.Node()
	wh := wallet.NewWalletHelperFromWallet(client, nep6Wallet)
	balancesGas, err := wh.GetAccountAndBalance(tx.GasToken)
	if err != nil {
		err = fmt.Errorf("WalletHelper.GetAccountAndBalance err %s", err)
		return
	}
	trx, err := wh.MakeTransaction(script, nil, []tx.ITransactionAttribute{}, balancesGas)
	if err != nil {
		err = fmt.Errorf("WalletHelper.MakeTransaction err %s", err)
		return
	}
	// sign transaction
	trx, err = wh.SignTransaction(trx, base.Neo3MagicNum)
	if err != nil {
		err = fmt.Errorf("WalletHelper.SignTransaction err %s", err)
		return
	}
	rawTxString := crypto.Base64Encode(trx.ToByteArray())
	// send the raw transaction
	res := client.SendRawTransaction(rawTxString)
	if res.HasError() {
		err = fmt.Errorf("send neo3 raw transaction err %s", res.ErrorResponse.Error.Message)
		return
	}
	hash = trx.GetHash().String()
	log.Info("Send neo3 transaction", "hash", hash)
	return
}

func (w *Neo3Wallet) Init() (err error) {
	return
}
