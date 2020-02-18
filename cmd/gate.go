package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/MinterTeam/explorer-gate/v2/api"
	"github.com/MinterTeam/explorer-gate/v2/core"
	sdk "github.com/MinterTeam/minter-go-sdk/api"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/libs/pubsub"
	"log"
	"os"
	"strconv"
	"time"
)

var Version string   // Version
var GitCommit string // Git commit
var BuildDate string // Build date
var AppName string   // Application name

var version = flag.Bool(`v`, false, `Prints current version`)

func main() {
	flag.Parse()
	if *version {
		fmt.Printf(`%s v%s Commit %s builded %s`, AppName, Version, GitCommit, BuildDate)
		os.Exit(0)
	}

	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found")
	}

	//Init Logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetReportCaller(true)
	if os.Getenv("GATE_DEBUG") != "1" {
		logger.SetFormatter(&logrus.TextFormatter{
			DisableColors: false,
			FullTimestamp: true,
		})
	} else {
		logger.SetFormatter(&logrus.JSONFormatter{})
		logger.SetLevel(logrus.WarnLevel)
	}

	contextLogger := logger.WithFields(logrus.Fields{
		"version": "2.0.0",
		"app":     "Minter Gate",
	})

	pubsubServer := pubsub.NewServer()
	err = pubsubServer.Start()
	if err != nil {
		contextLogger.Error(err)
	}

	gateService := core.New(pubsubServer, contextLogger)

	nodeApi := sdk.NewApi(os.Getenv("NODE_API"))

	status, err := nodeApi.Status()
	if err != nil {
		panic(err)
	}

	latestBlock, err := strconv.Atoi(status.LatestBlockHeight)
	if err != nil {
		panic(err)
	}

	logger.Info("Starting with block " + strconv.Itoa(latestBlock))

	go func() {
		for {
			block, err := nodeApi.Block(latestBlock)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}

			for _, tx := range block.Transactions {
				b, _ := hex.DecodeString(tx.RawTx)
				err := pubsubServer.PublishWithTags(context.TODO(), "NewTx", map[string]string{
					"tx": fmt.Sprintf("%X", b),
				})
				if err != nil {
					logger.Error(err)
				}
			}

			latestBlock++
			time.Sleep(1 * time.Second)
		}
	}()

	api.Run(gateService, pubsubServer)
}