// blacklist_xlayer.go
package txpool

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/ledgerwatch/erigon/rlp"
	"github.com/ledgerwatch/log/v3"

	"github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/types"
)

// IsTransferFromForBlockedAddress checks if a transaction is a transferFrom call with a blocked address as the from parameter
func IsTransferFromForBlockedAddress(txn *types.TxSlot, blockedList common.OrderedList[common.Address]) bool {
	log.Debug("TX TRACING: Full RLP", "rlp", fmt.Sprintf("%x", txn.Rlp))

	if txn.Creation || txn.To == (common.Address{}) {
		return false
	}

	transferFromSig := []byte{0x23, 0xb8, 0x72, 0xdd}

	data, err := getTxData(txn)
	if err != nil {
		return false
	}

	if len(data) < 4 {
		return false
	}

	methodID := data[:4]
	log.Debug("TX TRACING: Method ID", "methodID", fmt.Sprintf("%x", methodID))

	if !bytes.Equal(methodID, transferFromSig) {
		return false
	}

	if len(data) < 36 {
		return false
	}

	fromParam := common.BytesToAddress(data[4+12 : 4+32])
	log.Debug("TX TRACING: From Parameter", "fromParam", fmt.Sprintf("%x", fromParam))

	return blockedList.Contains(fromParam)
}

// getTxData extracts the data field from a transaction based on its type
func getTxData(tx *types.TxSlot) ([]byte, error) {
	switch tx.Type {
	case 0x00: // Legacy Transaction
		var txFields []interface{}
		if err := rlp.DecodeBytes(tx.Rlp, &txFields); err != nil {
			log.Debug("TX TRACING: Failed to decode legacy transaction", "error", err)
			return nil, err
		}
		if len(txFields) < 6 {
			log.Debug("TX TRACING: Invalid RLP data for legacy transaction", "fields", len(txFields))
			return nil, errors.New("invalid RLP data")
		}
		data, ok := txFields[5].([]byte)
		if !ok {
			log.Debug("TX TRACING: No valid data field in legacy transaction")
			return nil, errors.New("no valid data field")
		}
		return data, nil
	case 0x02: // EIP-1559
		var txFields []interface{}
		if err := rlp.DecodeBytes(tx.Rlp[1:], &txFields); err != nil {
			log.Debug("TX TRACING: Failed to decode EIP-1559 transaction", "error", err)
			return nil, err
		}
		if len(txFields) < 8 {
			log.Debug("TX TRACING: Invalid RLP data for EIP-1559 transaction", "fields", len(txFields))
			return nil, errors.New("invalid RLP data")
		}
		data, ok := txFields[7].([]byte)
		if !ok {
			log.Debug("TX TRACING: No valid data field in EIP-1559 transaction")
			return nil, errors.New("no valid data field")
		}
		return data, nil
	default:
		log.Debug("TX TRACING: Unsupported transaction type", "type", tx.Type)
		return nil, fmt.Errorf("unsupported tx type: %d", tx.Type)
	}
}
