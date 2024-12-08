package chaincode

import (
	"encoding/json"
	"fmt"
	"time"

	// "github.com/google/uuid"
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// Asset represents the asset structure
type Asset struct {
	IdOperasi       string `json:"IdOperasi"`
	NoResi          string `json:"noResi"`
	StatusPengiriman string `json:"statusPengiriman"`
	LokasiBarang    string `json:"lokasiBarang"`
	Operator        string `json:"operator"`
	BuktiStatus     string `json:"buktiStatus"`
	Timestamp       string `json:"timestamp"`
}

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// InitLedger initializes the ledger with some sample assets
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{
			IdOperasi:       "0001",
			NoResi:          "RESI001",
			StatusPengiriman: "Dikirim",
			LokasiBarang:    "Jakarta",
			Operator:        "JNE",
			BuktiStatus:     "Foto1.jpg",
			Timestamp:       time.Now().UTC().Format(time.RFC3339),
		},
		{
			IdOperasi:       "0002",
			NoResi:          "RESI002",
			StatusPengiriman: "Dalam Perjalanan",
			LokasiBarang:    "Bandung",
			Operator:        "SiCepat",
			BuktiStatus:     "Foto2.jpg",
			Timestamp:       time.Now().UTC().Format(time.RFC3339),
		},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return fmt.Errorf("failed to serialize asset: %v", err)
		}

		err = ctx.GetStub().PutState(asset.IdOperasi, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to store asset in world state: %v", err)
		}
	}

	return nil
}

// CreateAsset adds a new asset to the ledger
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, idOperasi, noResi, statusPengiriman, lokasiBarang, operator, buktiStatus string) error {
	// Generate unique ID for the asset
	timestamp := time.Now().UTC().Format(time.RFC3339)

	asset := Asset{
		IdOperasi:       idOperasi,
		NoResi:          noResi,
		StatusPengiriman: statusPengiriman,
		LokasiBarang:    lokasiBarang,
		Operator:        operator,
		BuktiStatus:     buktiStatus,
		Timestamp:       timestamp,
	}

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf("failed to serialize asset: %v", err)
	}

	err = ctx.GetStub().PutState(idOperasi, assetJSON)
	if err != nil {
		return fmt.Errorf("failed to store asset in world state: %v", err)
	}

	return nil
}

// GetAllAssets retrieves all assets from the ledger
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get assets: %v", err)
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate assets: %v", err)
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal asset: %v", err)
		}

		assets = append(assets, &asset)
	}

	return assets, nil
}
