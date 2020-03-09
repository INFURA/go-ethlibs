package main

import (
	"context"
	"flag"
	"github.com/INFURA/go-ethlibs/eth"
	"github.com/INFURA/go-ethlibs/jsonrpc"
	"github.com/INFURA/go-ethlibs/node"
	"log"
)

var (
	target = flag.String("target", "", "target URL")
)

func main() {
	flag.Parse()
	target := *target
	if target == "" {
		log.Fatalf("[FATAL] Target is required")
		return
	}

	ctx := context.Background()
	log.Printf("[INFO] Connecting to: %s", target)

	client, err := node.NewClient(ctx, target)
	if err != nil {
		log.Fatalf("[FATAL] Client error: %v", err)
	}

	num, err := client.BlockNumber(ctx)
	if err != nil {
		log.Fatalf("[FATAL] BlockNumber error: %v", err)
	}

	log.Printf("[INFO] Current Block number: %d", num)

	if client.IsBidirectional() {
		log.Printf("[INFO] Client supports subscriptions")
		log.Printf("[INFO] starting newHeads subscription...")
		newHeads, err := client.SubscribeNewHeads(ctx)
		if err != nil {
			log.Fatalf("[FATAL] NewHeads error: %v", err)
		}

		log.Printf("[INFO] newHeads subscription id %s", newHeads.ID())
		log.Printf("[INFO] waiting for newHeads subscription notifications...")
		received := 0
		for notif := range newHeads.Ch() {
			newHead := eth.NewHeadsNotificationParams{}
			err := notif.UnmarshalParamsInto(&newHead)
			if err != nil {
				log.Fatalf("[FATAL] Cannot parse newHeads params: %v", err)
			}

			// get the full block details, using a custom jsonrpc ID as a test
			block, err := client.BlockByHash(
				ctx, newHead.Result.Hash.String(),
				true,
				node.WithRequestID(jsonrpc.StringID("foo")),
			)
			if err != nil {
				log.Fatalf("[FATAL] Block for newHead notification not found: %v", err)
			}

			log.Printf("[INFO] got notification for new block %v %v", block.Number.UInt64(), block.Hash)

			received += 1
			// unsubscribe after 3rd block (just as a test)
			if received == 3 {
				if err := newHeads.Unsubscribe(ctx); err != nil {
					log.Fatalf("[FATAL] error unsubscribing newHeads: %v", err)
				}
			}
		}

		// we'll do a logs subscription as soon as the newHeads one is done
		log.Printf("[INFO] starting logs subscription...")
		logs, err := client.Subscribe(ctx, &jsonrpc.Request{
			JSONRPC: "2.0",
			Method:  "eth_subscribe",
			ID: jsonrpc.ID{
				Num: 2,
			},
			Params: jsonrpc.MustParams("logs", &eth.LogFilter{}),
		})
		if err != nil {
			log.Fatalf("[FATAL] Logs subscription error: %v", err)
		}

		log.Printf("[INFO] logs subscription id %s", logs.ID())
		log.Printf("[INFO] waiting for logs subscription notification...")
		for notif := range logs.Ch() {
			log.Printf("[INFO] Logs result: %s", string(notif.Params))

			// unsubscribe after first logs (just as a test)
			if err := logs.Unsubscribe(ctx); err != nil {
				log.Fatalf("[FATAL] error unsubscribing logs: %v", err)
			}
		}
	}
}
