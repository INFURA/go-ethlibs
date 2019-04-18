package main

import (
	"context"
	"flag"
	"github.com/INFURA/go-ethlibs/eth"
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

	if client.SupportsSubscriptions() {
		log.Printf("[INFO] Client should support subscriptions, attempting newHeads")
		newHeads, err := client.NewHeads(ctx)
		if err != nil {
			log.Fatalf("[FATAL] NewHeads error: %v", err)
		}

		for notif := range newHeads.Ch() {
			log.Printf("[INFO] %s", string(notif.Params))
			newHead := eth.NewHeadsNotificationParams{}
			err := notif.UnmarshalParamsInto(&newHead)
			if err != nil {
				log.Fatalf("[FATAL] Cannot parse newHeads params: %v", err)
			}

			log.Printf("[INFO] got notification for newHead %v %v", newHead.Result.Number.UInt64(), newHead.Result.Hash)
			return
		}
	}
}
