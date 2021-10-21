package chaincode

import (
	"encoding/json"
	"fabric-test/encryption"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/yilisita/goNum"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

var T = encryption.GetRandomMatrix(1, 1, 100)
var S = encryption.GetSecretKey(T)
// Asset describes basic details of what makes up a simple asset
//Insert struct field in alphabetic order => to achieve determinism accross languages
// golang keeps the order when marshal to json but doesn't order automatically
type Asset struct {
	UserName string `json:"UserName"`
	UserId   string `json:"UserId"`
	Amount   []float64   `json:"Amount"`  //2 * 1 的向量
}

func NewAsset(userName, userId string, amount float64) Asset{
	var a = Asset{}
	a.UserName = userName
	a.UserId = userId
	var aSlice = []float64{amount}
	var amountM = goNum.NewMatrix(1, 1, aSlice)
	a.Amount = encryption.Encrypt(T, amountM).Data
	return a

}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		NewAsset("xiao ming","1", 21),
		NewAsset("xiao hong", "2", 55),
		NewAsset("xiao liang", "3", 15),
		NewAsset("xiao hua",  "4",  37),
		NewAsset("xiao fei", "5", 49),
		NewAsset("xiao zhang", "6",  29),
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.UserId, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, userName string, userId string, amount float64) error {
	exists, err := s.AssetExists(ctx, userId)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", userId)
	}

	asset := NewAsset(userName, userId, amount)
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(userId, assetJSON)
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, userId string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", userId)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	asset.Amount = encryption.Decrypt(S, goNum.NewMatrix(2, 1, asset.Amount)).Data

	return &asset, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, userName string, userId string, amount float64) error {
	exists, err := s.AssetExists(ctx, userId)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", userId)
	}

	// overwriting original asset with new asset
	asset := NewAsset(userName, userId, amount)
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(userId, assetJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, userId string) error {
	exists, err := s.AssetExists(ctx, userId)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", userId)
	}

	return ctx.GetStub().DelState(userId)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, userId string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(userId)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

//// TransferAsset updates the owner field of asset with given id in world state.
//func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, id string, newOwner string) error {
//	asset, err := s.ReadAsset(ctx, id)
//	if err != nil {
//		return err
//	}
//
//	asset.Owner = newOwner
//	assetJSON, err := json.Marshal(asset)
//	if err != nil {
//		return err
//	}
//
//	return ctx.GetStub().PutState(id, assetJSON)
//}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, float64, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, -1, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	var res float64
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, -1, err
	}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, -1, err
		}
		asset.Amount = encryption.Decrypt(S, goNum.NewMatrix(2, 1, asset.Amount)).Data
		assets = append(assets, &asset)
		res += asset.Amount[0]
	}

	return assets, res, nil
}
