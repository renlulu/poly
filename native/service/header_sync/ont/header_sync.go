/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package ont

import (
	"fmt"

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/core/genesis"
	"github.com/ontio/multi-chain/core/types"
	"github.com/ontio/multi-chain/native"
	hscommon "github.com/ontio/multi-chain/native/service/header_sync/common"
	"github.com/ontio/multi-chain/native/service/utils"
	ocommon "github.com/ontio/ontology/common"
	otypes "github.com/ontio/ontology/core/types"
	bcommon "github.com/ontio/ontology/http/base/common"
)

type ONTHandler struct {
}

func NewONTHandler() *ONTHandler {
	return &ONTHandler{}
}

func (this *ONTHandler) SyncGenesisHeader(native *native.NativeService) error {
	params := new(hscommon.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("SyncGenesisHeader, contract params deserialize error: %v", err)
	}

	// get operator from database
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return err
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return fmt.Errorf("SyncGenesisHeader, checkWitness error: %v", err)
	}

	header, err := otypes.HeaderFromRawBytes(params.GenesisHeader)
	if err != nil {
		return fmt.Errorf("SyncGenesisHeader, deserialize header err: %v", err)
	}
	//block header storage
	err = PutBlockHeader(native, params.ChainID, header)
	if err != nil {
		return fmt.Errorf("SyncGenesisHeader, put blockHeader error: %v", err)
	}

	//consensus node pk storage
	err = UpdateConsensusPeer(native, params.ChainID, header)
	if err != nil {
		return fmt.Errorf("SyncGenesisHeader, update ConsensusPeer error: %v", err)
	}
	return nil
}

func (this *ONTHandler) SyncBlockHeader(native *native.NativeService) error {
	params := new(hscommon.SyncBlockHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("SyncBlockHeader, contract params deserialize error: %v", err)
	}
	for _, v := range params.Headers {
		header, err := otypes.HeaderFromRawBytes(v)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, otypes.HeaderFromRawBytes error: %v", err)
		}
		_, err = GetHeaderByHeight(native, params.ChainID, header.Height)
		if err == nil {
			return fmt.Errorf("SyncBlockHeader, header already synced, %d, %d", params.ChainID, header.Height)
		}
		err = verifyHeader(native, params.ChainID, header)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, verifyHeader error: %v", err)
		}
		err = PutBlockHeader(native, params.ChainID, header)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, put BlockHeader error: %v", err)
		}
		err = UpdateConsensusPeer(native, params.ChainID, header)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, update ConsensusPeer error: %v", err)
		}
	}
	return nil
}

func (this *ONTHandler) SyncCrossChainMsg(native *native.NativeService) error {
	params := new(hscommon.SyncCrossChainMsgParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("SyncCrossChainMsg, contract params deserialize error: %v", err)
	}
	for _, v := range params.CrossChainMsgs {
		crossChainMsg := new(bcommon.CrossChainMsg)
		err := crossChainMsg.Deserialization(ocommon.NewZeroCopySource(v))
		if err != nil {
			return fmt.Errorf("SyncCrossChainMsg, deserialize crossChainMsg error: %v", err)
		}
		_, err = GetCrossChainMsg(native, params.ChainID, crossChainMsg.Height)
		if err == nil {
			return fmt.Errorf("SyncCrossChainMsg, crossChainMsg already synced, %d, %d", params.ChainID,
				crossChainMsg.Height)
		}
		err = VerifyCrossChainMsg(native, params.ChainID, crossChainMsg)
		if err != nil {
			return fmt.Errorf("SyncCrossChainMsg, VerifyCrossChainMsg error: %v", err)
		}
		err = PutCrossChainMsg(native, params.ChainID, crossChainMsg)
		if err != nil {
			return fmt.Errorf("SyncCrossChainMsg, put PutCrossChainMsg error: %v", err)
		}
	}
	return nil
}
