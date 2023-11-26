package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/golang/protobuf/ptypes"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/files"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const (
	channelName    = "mychannel"
	contractName   = "myassetchaincode"
	connectionFile = "connection.yaml"
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

// @title My Asset Chaincode API
// @version 1.0
// @description API for managing assets using Hyperledger Fabric Chaincode
// @host localhost:8080
// @BasePath /v1
func main() {
	r := gin.Default()

	// Setup Fabric Gateway
	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(connectionFile)),
		gateway.WithIdentity(&gateway.X509Identity{}),
	)
	if err != nil {
		fmt.Printf("Failed to connect to gateway: %s\n", err)
		return
	}
	defer gw.Close()

	network, err := gw.GetNetwork(channelName)
	if err != nil {
		fmt.Printf("Failed to get network: %s\n", err)
		return
	}

	contract := network.GetContract(contractName)

	// Create Asset Endpoint
	// @Summary Create an asset
	// @Description Create a new asset with the provided details
	// @Accept json
	// @Produce json
	// @Param input body Asset true "Asset details"
	// @Success 200 {string} string "Asset created successfully"
	// @Failure 400 {object} string "Bad Request"
	// @Failure 500 {object} string "Internal Server Error"
	// @Router /createAsset [post]
	r.POST("/createAsset", func(c *gin.Context) {
		var asset Asset
		if err := c.ShouldBindJSON(&asset); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Invoke Fabric Chaincode
		_, err := contract.SubmitTransaction("CreateAsset", asset.DealerID, asset.MSISDN, asset.MPIN, asset.Balance, asset.Status, asset.TransType, asset.Remarks)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Asset created successfully"})
	})

	// Update Asset Endpoint
	// @Summary Update an asset
	// @Description Update an existing asset with the provided details
	// @Accept json
	// @Produce json
	// @Param msisdn path string true "MSISDN of the asset to update"
	// @Param input body Asset true "Updated asset details"
	// @Success 200 {string} string "Asset updated successfully"
	// @Failure 400 {object} string "Bad Request"
	// @Failure 500 {object} string "Internal Server Error"
	// @Router /updateAsset/{msisdn} [post]
	r.POST("/updateAsset/:msisdn", func(c *gin.Context) {
		var asset Asset
		if err := c.ShouldBindJSON(&asset); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		msisdn := c.Param("msisdn")

		// Invoke Fabric Chaincode
		_, err := contract.SubmitTransaction("UpdateAsset", msisdn, asset.Balance, asset.Status, asset.TransType, asset.Remarks)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Asset updated successfully"})
	})

	// Read Asset Endpoint
	// @Summary Read asset details
	// @Description Get details of an asset by MSISDN
	// @Produce json
	// @Param msisdn path string true "MSISDN of the asset to get details"
	// @Success 200 {object} Asset "Asset details"
	// @Failure 400 {object} string "Bad Request"
	// @Failure 500 {object} string "Internal Server Error"
	// @Router /readAsset/{msisdn} [get]
	r.GET("/readAsset/:msisdn", func(c *gin.Context) {
		msisdn := c.Param("msisdn")

		// Invoke Fabric Chaincode
		response, err := contract.EvaluateTransaction("ReadAsset", msisdn)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var asset Asset
		if err := json.Unmarshal(response, &asset); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, asset)
	})

	// Get Asset History Endpoint
	// @Summary Get asset history
	// @Description Get transaction history of an asset by MSISDN
	// @Produce json
	// @Param msisdn path string true "MSISDN of the asset to get history"
	// @Success 200 {array} AssetHistoryEntry "Transaction history"
	// @Failure 400 {object} string "Bad Request"
	// @Failure 500 {object} string "Internal Server Error"
	// @Router /getAssetHistory/{msisdn} [get]
	r.GET("/getAssetHistory/:msisdn", func(c *gin.Context) {
		msisdn := c.Param("msisdn")

		// Invoke Fabric Chaincode
		response, err := contract.EvaluateTransaction("GetAssetHistory", msisdn)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var historyRes []*AssetHistoryEntry
		if err := json.Unmarshal(response, &historyRes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, historyRes)
	})

	// Swagger documentation routes
	// @router /swagger/*any [get]
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Run the REST API
	err = r.Run(":8080")
	if err != nil {
		fmt.Printf("Failed to start REST API: %s\n", err)
	}
}
