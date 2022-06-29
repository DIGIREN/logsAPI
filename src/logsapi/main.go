package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/DIGIREN/logsAPI/cache"
	"github.com/DIGIREN/logsAPI/logging"
	"github.com/DIGIREN/logsAPI/models"
	"github.com/gin-gonic/gin"
)

//Setting up global logger for this file
var logger = logging.GetLogger()
var batchListenerInterval int

//Responsible for starting our server, batch listener, and setting up our endpoints
func main() {
	logging.InitLogger()
	//Re-setting logger after initilization
	logger = logging.GetLogger()

	server := gin.Default()
	trustedProxy := initRequiredVar("TRUSTED_PROXY", "localhost")
	server.SetTrustedProxies([]string{trustedProxy})

	go startBatchListener()
	logger.Info("Started batch listener")

	server.POST("/log", func(ctx *gin.Context) {
		var log models.LogObject
		ctx.BindJSON(&log)
		err := storeLogInCache(log)
		if err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			logger.Warnw("Failed to store log",
				"error", err.Error(),
			)
			return
		} else {
			ctx.JSON(200, gin.H{"success": "Log successfully stored"})
		}
		err = sendBatchIfSizeMet()
		if err != nil {
			logger.Warnw("Failed to send batch",
				"error", err.Error(),
			)
		}
	})

	server.GET("/healthz", func(ctx *gin.Context) {
		ctx.Header("Content-Type", "text/plain")
		ctx.String(200, "OK")
	})

	logger.Infof("Starting server on port %s", os.Getenv("PORT"))
	server.Run(":" + os.Getenv("PORT"))
}

//Stores a log in memory
func storeLogInCache(log models.LogObject) error {
	return cache.CacheLog(log)
}

//Makes a POST request to POST_ENDPOINT with JSON containing the current log batch
//If the request fails, it will retry a specified number of times accouring to MAX_RETRIES env var
//RETRY_WAIT will determine the amount of time to wait between retries
//POST_ENDPOINT will determine the remote endpoint to send our log batches to
func sendAllLogsToEndpoint() (float64, int, error) {
	postEndpoint := initRequiredVar("POST_ENDPOINT", "")
	logs := cache.GetAllLogs()

	//Build the JSON payload
	jsonData, err := json.Marshal(logs)
	if err != nil {
		return 0, 500, errors.New(err.Error())
	}

	retryCounter := 0
	maxRetries := initRequiredIntVar("MAX_RETRIES", "3")
	retryWait := initRequiredIntVar("RETRY_WAIT", "2")

	//Try to send the logs to the endpoint, if it fails, log and exit
	requestStart := time.Now()
	for retryCounter < maxRetries {
		response, err := sendJsonPayload(postEndpoint, &jsonData)
		logger.Info("Sending logs to endpoint:", postEndpoint, ". Attempt", retryCounter+1, "/", maxRetries)
		if err != nil {
			retryCounter++
			logger.Warnf("Failed with error:", err.Error(), "Retrying in", retryWait, "seconds")
			time.Sleep(time.Duration(retryWait) * time.Second)
			continue
		}
		if response.StatusCode == 200 {
			requestEnd := time.Since(requestStart)
			return requestEnd.Seconds(), response.StatusCode, nil
		}
	}
	return 0, 500, errors.New("Failed to send logs to " + postEndpoint)
}

//Sends a JSON payload of logs to a remote endpoint
func sendJsonPayload(url string, jsonData *[]byte) (*http.Response, error) {
	response, err := http.Post(url, "application/json", bytes.NewBuffer(*jsonData))
	if err != nil {
		return nil, errors.New(err.Error())
	}
	if response.StatusCode != 200 {
		return nil, errors.New("Failed to send logs to " + url + " with status code " + response.Status)
	}
	return response, nil
}

//Logs the batch size, result status code, and duration of the POST request to the external endpoint
func logRequestDetails(batchSize int, statusCode int, duration float64) {
	logger.Infow("Successfully sent log batch to endpoint",
		"batch_size", batchSize,
		"status_code", statusCode,
		"duration", duration,
	)
}

//Stops the server by throwing a fatal log
func terminateServer() {
	logger.Fatal("Remote endpoint is not available, failed after maxRetries was hit, terminating server")
}

//Starts a loop that will send logs to the remote endpoint when our current interval exceeds the batch size
func startBatchListener() error {
	checkFrequency := initRequiredIntVar("CHECK_FREQUENCY", "10")
	maxInterval := initRequiredIntVar("BATCH_INTERVAL", "10")
	batchSize := initRequiredIntVar("BATCH_SIZE", "20")
	for range time.Tick(time.Second * time.Duration(checkFrequency)) {
		batchListenerInterval += checkFrequency
		currentBatchSize := len(cache.GetAllLogs())
		if len(cache.GetAllLogs()) > 0 {
			if batchListenerInterval >= maxInterval || batchSize == currentBatchSize {
				sendBatch()
			}
		}
	}
	return nil
}

//Fires when a new log is added to the cache to see if its time to send a batch to our POST endpoint
func sendBatchIfSizeMet() error {
	batchSize := initRequiredIntVar("BATCH_SIZE", "20")
	currentBatchSize := len(cache.GetAllLogs())
	if currentBatchSize >= batchSize {
		sendBatch()
	}
	return nil
}

func sendBatch() error {
	duration, responseCode, err := sendAllLogsToEndpoint()
	if err != nil {
		terminateServer()
		return err
	} else {
		logRequestDetails(len(cache.GetAllLogs()), responseCode, duration)
	}
	batchListenerInterval = 0
	cache.EmptyAll()
	return nil
}

//Will load a required env var as a string or int
//will fail critically if these are not set & no defaults are passed, as they are required for runtime
//varString (string) - the name of the env var to load
//defaultVal (string) the default value to use if the env var is not set
//returns the string value of the env var
func initRequiredVar(varString string, defaultVal string) string {
	var envValue string

	if os.Getenv(varString) == "" {
		if len(defaultVal) > 0 {
			logger.Infof(varString, "is not set, using default value", defaultVal[0])
			envValue = defaultVal
		} else {
			logger.Fatal(varString, "is not set, and invalid default value was not provided")
		}
	} else {
		envValue = os.Getenv(varString)
	}
	return envValue
}

//Will load a required env var and then try to convert it to an int
func initRequiredIntVar(varString string, defaultVal string) int {
	envValue := initRequiredVar(varString, defaultVal)
	intVal, err := strconv.Atoi(envValue)
	if err != nil {
		logger.Fatal(varString, "is not a valid integer")
	}
	return intVal
}
