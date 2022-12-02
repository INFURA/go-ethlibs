package metrics

import (
	"context"
	"time"

	"github.com/ConsenSys/infura-binance/pkg/conf"
	"github.com/ConsenSys/infura-binance/pkg/jsonrpc"
	"go.uber.org/zap"
)

// Recorder struct.
type Recorder struct {
	Config   conf.HealthCheckEnvSpec
	Logger   *zap.SugaredLogger
	Client   *jsonrpc.Client
	Interval time.Duration
}

// NewRecorder create a new metrics recorder.
func NewRecorder(
	config conf.HealthCheckEnvSpec,
	logger *zap.SugaredLogger,
	client *jsonrpc.Client,
	interval time.Duration,
) *Recorder {
	return &Recorder{
		Config:   config,
		Logger:   logger,
		Client:   client,
		Interval: interval,
	}
}

// Start records metrics in a regular interval.
func (r *Recorder) Start(ctx context.Context) {
	go func() {
		go r.RecordClientHeadBlockNumber(ctx, r.Config.ClientEndpoint)

		// Stagger
		time.Sleep(1 * time.Second)
	}()
}
