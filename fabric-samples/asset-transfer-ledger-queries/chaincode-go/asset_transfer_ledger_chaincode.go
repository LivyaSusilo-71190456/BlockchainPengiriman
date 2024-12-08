/*
 SPDX-License-Identifier: Apache-2.0
*/

/*
====CHAINCODE EXECUTION SAMPLES (CLI) ==================

==== Invoke assets ====
peer chaincode invoke -C myc1 -n asset_transfer -c '{"Args":["CreateAsset","asset1","blue","5","tom","35"]}'
peer chaincode invoke -C myc1 -n asset_transfer -c '{"Args":["CreateAsset","asset2","red","4","tom","50"]}'
peer chaincode invoke -C myc1 -n asset_transfer -c '{"Args":["CreateAsset","asset3","blue","6","tom","70"]}'
peer chaincode invoke -C myc1 -n asset_transfer -c '{"Args":["TransferAsset","asset2","jerry"]}'
peer chaincode invoke -C myc1 -n asset_transfer -c '{"Args":["TransferAssetByColor","blue","jerry"]}'
peer chaincode invoke -C myc1 -n asset_transfer -c '{"Args":["DeleteAsset","asset1"]}'

==== Query assets ====
peer chaincode query -C myc1 -n asset_transfer -c '{"Args":["ReadAsset","asset1"]}'
peer chaincode query -C myc1 -n asset_transfer -c '{"Args":["GetAssetsByRange","asset1","asset3"]}'
peer chaincode query -C myc1 -n asset_transfer -c '{"Args":["GetAssetHistory","asset1"]}'

Rich Query (Only supported if CouchDB is used as state database):
peer chaincode query -C myc1 -n asset_transfer -c '{"Args":["QueryAssetsByOwner","tom"]}'
peer chaincode query -C myc1 -n asset_transfer -c '{"Args":["QueryAssets","{\"selector\":{\"owner\":\"tom\"}}"]}'

Rich Query with Pagination (Only supported if CouchDB is used as state database):
peer chaincode query -C myc1 -n asset_transfer -c '{"Args":["QueryAssetsWithPagination","{\"selector\":{\"owner\":\"tom\"}}","3",""]}'

INDEXES TO SUPPORT COUCHDB RICH QUERIES

Indexes in CouchDB are required in order to make JSON queries efficient and are required for
any JSON query with a sort. Indexes may be packaged alongside
chaincode in a META-INF/statedb/couchdb/indexes directory. Each index must be defined in its own
text file with extension *.json with the index definition formatted in JSON following the
CouchDB index JSON syntax as documented at:
http://docs.couchdb.org/en/2.3.1/api/database/find.html#db-index

This asset transfer ledger example chaincode demonstrates a packaged
index which you can find in META-INF/statedb/couchdb/indexes/indexOwner.json.

If you have access to the your peer's CouchDB state database in a development environment,
you may want to iteratively test various indexes in support of your chaincode queries.  You
can use the CouchDB Fauxton interface or a command line curl utility to create and update
indexes. Then once you finalize an index, include the index definition alongside your
chaincode in the META-INF/statedb/couchdb/indexes directory, for packaging and deployment
to managed environments.

In the examples below you can find index definitions that support asset transfer ledger
chaincode queries, along with the syntax that you can use in development environments
to create the indexes in the CouchDB Fauxton interface or a curl command line utility.


Index for docType, owner.

Example curl command line to define index in the CouchDB channel_chaincode database
curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[\"docType\",\"owner\"]},\"name\":\"indexOwner\",\"ddoc\":\"indexOwnerDoc\",\"type\":\"json\"}" http://hostname:port/myc1_assets/_index


Index for docType, owner, size (descending order).

Example curl command line to define index in the CouchDB channel_chaincode database:
curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[{\"size\":\"desc\"},{\"docType\":\"desc\"},{\"owner\":\"desc\"}]},\"ddoc\":\"indexSizeSortDoc\", \"name\":\"indexSizeSortDesc\",\"type\":\"json\"}" http://hostname:port/myc1_assets/_index

Rich Query with index design doc and index name specified (Only supported if CouchDB is used as state database):
peer chaincode query -C myc1 -n asset_transfer -c '{"Args":["QueryAssets","{\"selector\":{\"docType\":\"asset\",\"owner\":\"tom\"}, \"use_index\":[\"_design/indexOwnerDoc\", \"indexOwner\"]}"]}'

Rich Query with index design doc specified only (Only supported if CouchDB is used as state database):
peer chaincode query -C myc1 -n asset_transfer -c '{"Args":["QueryAssets","{\"selector\":{\"docType\":{\"$eq\":\"asset\"},\"owner\":{\"$eq\":\"tom\"},\"size\":{\"$gt\":0}},\"fields\":[\"docType\",\"owner\",\"size\"],\"sort\":[{\"size\":\"desc\"}],\"use_index\":\"_design/indexSizeSortDoc\"}"]}'
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	// "github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

const index = "noResi"

// SimpleChaincode implements the fabric-contract-api-go programming model
type SimpleChaincode struct {
	contractapi.Contract
}

type Asset struct {
	NoResi          string `json:"noResi"`
	StatusPengiriman string `json:"statusPengiriman"`
	LokasiBarang    string `json:"lokasiBarang"`
	Operator        string `json:"operator"`
	BuktiStatus     string `json:"buktiStatus"`
	Timestamp       string `json:"timestamp"`
}

// HistoryQueryResult structure used for returning result of history query
type HistoryQueryResult struct {
	Record    *Asset    `json:"record"`
	TxId      string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

// CreateAsset initializes a new asset in the ledger
func (t *SimpleChaincode) CreateAsset(ctx contractapi.TransactionContextInterface, noResi, statusPengiriman, lokasiBarang, operator, buktiStatus string) error {
	exists, err := t.AssetExists(ctx, noResi)
	if err != nil {
		return fmt.Errorf("failed to check asset existence: %v", err)
	}
	if exists {
		return fmt.Errorf("asset with noResi %s already exists", noResi)
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)

	asset := &Asset{
		NoResi:          noResi,
		StatusPengiriman: statusPengiriman,
		LokasiBarang:    lokasiBarang,
		Operator:        operator,
		BuktiStatus:     buktiStatus,
		Timestamp:       timestamp,
	}

	assetBytes, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(noResi, assetBytes)
	if err != nil {
		return err
	}

	// Create composite key for index
	statusIndexKey, err := ctx.GetStub().CreateCompositeKey(index, []string{asset.StatusPengiriman, asset.NoResi})
	if err != nil {
		return err
	}

	// Save index entry to world state
	value := []byte{0x00}
	return ctx.GetStub().PutState(statusIndexKey, value)
}

// ReadAsset retrieves an asset from the ledger
func (t *SimpleChaincode) ReadAsset(ctx contractapi.TransactionContextInterface, noResi string) (*Asset, error) {
	assetBytes, err := ctx.GetStub().GetState(noResi)
	if err != nil {
		return nil, fmt.Errorf("failed to read asset %s: %v", noResi, err)
	}
	if assetBytes == nil {
		return nil, fmt.Errorf("asset with noResi %s does not exist", noResi)
	}

	var asset Asset
	err = json.Unmarshal(assetBytes, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// TransferAsset transfers an asset by setting a new operator
func (t *SimpleChaincode) TransferAsset(ctx contractapi.TransactionContextInterface, noResi, newStatusPengiriman, newLokasiBarang, newOperator, newBuktiStatus string) error {
	// Membaca asset berdasarkan noResi
	asset, err := t.ReadAsset(ctx, noResi)
	if err != nil {
		return err
	}

	// Memperbarui field dalam asset
	asset.StatusPengiriman = newStatusPengiriman
	asset.LokasiBarang = newLokasiBarang
	asset.Operator = newOperator
	asset.BuktiStatus = newBuktiStatus
	asset.Timestamp = time.Now().UTC().Format(time.RFC3339)

	// Mengubah asset menjadi byte slice untuk disimpan di ledger
	assetBytes, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	// Menyimpan asset yang telah diperbarui ke ledger
	return ctx.GetStub().PutState(noResi, assetBytes)
}

// AssetExists checks if an asset exists in the ledger
func (t *SimpleChaincode) AssetExists(ctx contractapi.TransactionContextInterface, noResi string) (bool, error) {
	assetBytes, err := ctx.GetStub().GetState(noResi)
	if err != nil {
		return false, fmt.Errorf("failed to read asset %s from world state. %v", noResi, err)
	}

	return assetBytes != nil, nil
}

func (t *SimpleChaincode) DeleteAsset(ctx contractapi.TransactionContextInterface, noResi string) error {
	asset, err := t.ReadAsset(ctx, assetID)
	if err != nil {
		return err
	}

	err = ctx.GetStub().DelState(assetID)
	if err != nil {
		return fmt.Errorf("failed to delete asset %s: %v", assetID, err)
	}

	colorNameIndexKey, err := ctx.GetStub().CreateCompositeKey(index, []string{asset.Color, asset.ID})
	if err != nil {
		return err
	}

	// Delete index entry
	return ctx.GetStub().DelState(colorNameIndexKey)
}

// QueryAssetsByOperator queries for assets by operator
func (t *SimpleChaincode) QueryAssetsByOperator(ctx contractapi.TransactionContextInterface, operator string) ([]*Asset, error) {
	queryString := fmt.Sprintf(`{"selector":{"operator":"%s"}}`, operator)
	return getQueryResultForQueryString(ctx, queryString)
}

// GetAssetHistory returns the chain of custody for an asset since issuance
func (t *SimpleChaincode) GetAssetHistory(ctx contractapi.TransactionContextInterface, noResi string) ([]HistoryQueryResult, error) {
	log.Printf("GetAssetHistory: noResi %v", noResi)

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(noResi)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &asset)
			if err != nil {
				return nil, err
			}
		} else {
			asset = Asset{
				NoResi: noResi,
			}
		}

		record := HistoryQueryResult{
			TxId:      response.TxId,
			Timestamp: response.Timestamp.AsTime(),
			Record:    &asset,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}

// Helper function for rich queries
func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*Asset, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset Asset
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

// GetAllAssets retrieves all assets from the ledger
func (t *SimpleChaincode) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
    // Retrieve all records using an empty key range
    resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
    if err != nil {
        return nil, fmt.Errorf("failed to get assets: %v", err)
    }
    defer resultsIterator.Close()

    var assets []*Asset
    for resultsIterator.HasNext() {
        queryResult, err := resultsIterator.Next()
        if err != nil {
            return nil, fmt.Errorf("failed to iterate over assets: %v", err)
        }

        var asset Asset
        err = json.Unmarshal(queryResult.Value, &asset)
        if err != nil {
            return nil, fmt.Errorf("failed to unmarshal asset %s: %v", queryResult.Key, err)
        }

        assets = append(assets, &asset)
    }

    return assets, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SimpleChaincode{})
	if err != nil {
		log.Panicf("Error creating chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting chaincode: %v", err)
	}
}
