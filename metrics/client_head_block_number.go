package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var clientHeadBlockNumber = promauto.NewGauge( //nolint:gochecknoglobals
	prometheus.GaugeOpts{ //nolint:exhaustivestruct,exhaustruct
		Name: "client_head_block_number",
		Help: "The head block number of the client",
	},
)

func (r *Recorder) RecordClientHeadBlockNumber(ctx context.Context) {
	ticker := time.NewTicker(r.Interval)

	for {
		<-ticker.C

		r.Logger.Debug("Metric sub-process:recordClientHeadBlockNumber is running")

		blockNumber, err := r.Client.BlockNumber(ctx)
		if err != nil {
			r.Logger.Errorf("metrics.clientHeadBlockNumber failed... %s", err)

			continue
		}

		clientHeadBlockNumber.Set(float64(blockNumber))

		r.Logger.Debugf("metrics.clientHeadBlockNumber passed. current: %v", blockNumber)
	}
}
