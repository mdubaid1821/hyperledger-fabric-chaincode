package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/golang/protobuf/ptypes"
)

// Asset describes the structure of an asset
type Asset struct {
	DealerID     string    `json:"DealerID"`
	MSISDN       string    `json:"MSISDN"`
	MPIN         string    `json:"MPIN"`
	Balance      int       `json:"Balance"`
	Status       string    `json:"Status"`
	TransAmount  int       `json:"TransAmount"`
	TransType    string    `json:"TransType"`
	Remarks      string    `json:"Remarks"`
	Timestamp    time.Time `json:"Timestamp"`
}

// AssetHistoryEntry describes an entry in the asset transaction history
type AssetHistoryEntry struct {
	TxID      string    `json:"TxID"`
	Timestamp time.Time `json:"Timestamp"`
}

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{DealerID: "D001", MSISDN: "1234567890", MPIN: "1234", Balance: 1000, Status: "Active", TransAmount: 0, TransType: "", Remarks: ""},
		{DealerID: "D002", MSISDN: "9876543210", MPIN: "5678", Balance: 1500, Status: "Active", TransAmount: 0, TransType: "", Remarks: ""},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.MSISDN, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state: %v", err)
		}
	}

	return nil
}

// CreateAsset creates a new asset and stores it on the ledger
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, dealerID, msisdn, mpin string, balance int, status, transType, remarks string) error {
	exists, err := s.AssetExists(ctx, msisdn)
	if err != nil {
		return fmt.Errorf("error checking asset existence: %v", err)
	}
	if exists {
		return fmt.Errorf("asset with MSISDN %s already exists", msisdn)
	}

	asset := Asset{
		DealerID:    dealerID,
		MSISDN:      msisdn,
		MPIN:        mpin,
		Balance:     balance,
		Status:      status,
		TransAmount:  0,
		TransType:   "",
		Remarks:     "",
	}

	// Get transaction timestamp
	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("error getting transaction timestamp: %v", err)
	}
	asset.Timestamp, err = ptypes.Timestamp(txTimestamp)
	if err != nil {
		return fmt.Errorf("error converting timestamp: %v", err)
	}

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf("error marshalling asset: %v", err)
	}

	return ctx.GetStub().PutState(msisdn, assetJSON)
}

// UpdateAsset updates the values of an existing asset
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, msisdn, newBalanceStr, newStatus, transType, remarks string) error {
    fmt.Printf("Received arguments: msisdn=%s, newBalanceStr=%s, newStatus=%s, transType=%s, remarks=%s\n", msisdn, newBalanceStr, newStatus, transType, remarks)

	asset, err := s.ReadAsset(ctx, msisdn)
	if err != nil {
		return fmt.Errorf("error reading asset: %v", err)
	}

	// Convert newBalanceStr to integer
	newBalance, err := strconv.Atoi(newBalanceStr)
	if err != nil {
        fmt.Printf("Error converting newBalanceStr to integer: %v\n", err)
		return fmt.Errorf("error converting newBalanceStr to integer: %v", err)
	}

	asset.Balance = newBalance
	asset.Status = newStatus
	asset.TransAmount = newBalance - asset.Balance
	asset.TransType = transType
	asset.Remarks = remarks

	// Get transaction timestamp
	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
        fmt.Printf("Error getting transaction timestamp: %v\n", err)
		return fmt.Errorf("error getting transaction timestamp: %v", err)
	}
	asset.Timestamp, err = ptypes.Timestamp(txTimestamp)
	if err != nil {
        fmt.Printf("Error converting timestamp: %v\n", err)
		return fmt.Errorf("error converting timestamp: %v", err)
	}

	assetJSON, err := json.Marshal(asset)
	if err != nil {
        fmt.Printf("Error marshalling asset: %v\n", err)
		return fmt.Errorf("error marshalling asset: %v", err)
	}

	return ctx.GetStub().PutState(msisdn, assetJSON)
}

// ReadAsset retrieves the current state of an asset
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, msisdn string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(msisdn)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("asset with MSISDN %s does not exist", msisdn)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling asset: %v", err)
	}

	return &asset, nil
}

// GetAssetHistory retrieves the transaction history of an asset
func (s *SmartContract) GetAssetHistory(ctx contractapi.TransactionContextInterface, msisdn string) ([]*AssetHistoryEntry, error) {
    resultsIterator, err := ctx.GetStub().GetHistoryForKey(msisdn)
    if err != nil {
        return nil, fmt.Errorf("error getting asset history: %v", err)
    }
    defer resultsIterator.Close()

    var history []*AssetHistoryEntry
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, fmt.Errorf("error iterating through history: %v", err)
        }

        var entry AssetHistoryEntry
        entry.TxID = queryResponse.TxId
        entry.Timestamp, err = ptypes.Timestamp(queryResponse.Timestamp)
        if err != nil {
            return nil, fmt.Errorf("error converting timestamp: %v", err)
        }

        history = append(history, &entry)
    }

    return history, nil
}



// AssetExists checks if an asset with the given MSISDN exists
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, msisdn string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(msisdn)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

func main() {
	assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		fmt.Printf("Error creating asset chaincode: %s", err.Error())
		return
	}

	if err := assetChaincode.Start(); err != nil {
		fmt.Printf("Error starting asset chaincode: %s", err.Error())
	}
}
