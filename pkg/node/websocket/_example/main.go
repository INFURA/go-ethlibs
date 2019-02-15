package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/INFURA/ethereum-interaction/pkg/eth"
	"github.com/INFURA/ethereum-interaction/pkg/node/websocket"
)

var endpoint = flag.String("URL", "wss://mainnet.infura.io/ws", "The websocket endpoint to connect to")
var verbose = flag.Bool("verbose", false, "if set newHeads and block JSON will be printed")

func main() {
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())

	url := *endpoint
	client, err := websocket.NewConnection(ctx, url)
	if err != nil {
		log.Fatalf("[FATAL] could not connect to %s: %v", url, err)
	}

	log.Printf("[INFO] Connected to %s", client.URL())

	blockNumber, err := client.BlockNumber(ctx)
	if err != nil {
		log.Fatalf("[FATAL] could not get block number: %v", err)
	}

	log.Printf("[INFO] Current block number: %v", blockNumber)

	latest, err := client.BlockByNumber(ctx, blockNumber, true)
	if err != nil {
		log.Fatalf("[FATAL] could not get block: %v", err)
	}

	log.Printf("[INFO] Latest block total difficulty: %s", latest.TotalDifficulty.Big().String())

	log.Printf("[INFO] Subscribing to new blocks")
	subscription, err := client.NewHeads(ctx)
	if err != nil {
		log.Fatalf("[FATAL] could not subscribe to newHeads: %v", err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	loop := true

	for loop {
		select {
		case notification := <-subscription.Ch():
			params := eth.NewHeadsNotificationParams{}
			err := json.Unmarshal(notification.Params, &params)
			if err != nil {
				log.Fatalf("[FATAL] could not decode notification params: %v", err)
			}

			if *verbose {
				block, err := client.BlockByHash(ctx, params.Result.Hash.String(), false)
				if err != nil {
					log.Fatalf("[FATAL] could not get latest block: %v", err)
				}

				b, err := json.Marshal(&block)
				if err != nil {
					log.Fatalf("[FATAL] could not marshal latest block: %v", err)
				}

				log.Printf("[INFO] Raw newHead: %s", string(notification.Params))
				log.Printf("[INFO] Block JSON: %s", string(b))

			} else {
				log.Printf("[INFO] New block %d hash: %s, parent: %s", params.Result.Number.UInt64(), params.Result.Hash.String(), params.Result.ParentHash.String())
			}

		case <-subscription.Done():
			log.Printf("[WARN] Disconnected: %v", subscription.Err())
			loop = false
			cancel()

		case <-ctx.Done():
			log.Printf("[INFO] Done")
			loop = false

		case sig := <-signals:
			// Print an empty line when we get the signal to keep the log pretty
			fmt.Println()

			switch sig {
			case syscall.SIGUSR2:
				// Log debug state
				log.Printf("[SIGUSR2] %v", client)
			default:
				// for any other signal sent, we want to shut down
				log.Printf("[INFO] Unsubscribing")
				err := subscription.Unsubscribe(ctx)
				if err != nil {
					log.Fatalf("[ERROR] error unsubscribing %v", err)
				}

				loop = false
				cancel()
			}
		}
	}

	log.Printf("[DEBUG] waiting for cleanup")
	<-ctx.Done()
}
