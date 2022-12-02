package metrics

import (
	"context"
	"time"

	"github.com/ConsenSys/go-ethlibs/node"
	"github.com/ConsenSys/go-ethlibs/conf"
	"go.uber.org/zap"
)

// Recorder struct.
type Recorder struct {
	Config   conf.HealthCheckEnvSpec
	Logger   *zap.SugaredLogger
	Client   node.Client
	Interval time.Duration
}

// NewRecorder create a new metrics recorder.
func NewRecorder(
	config conf.HealthCheckEnvSpec,
	logger *zap.SugaredLogger,
	client node.Client,
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
		go r.RecordClientHeadBlockNumber(ctx)

		// Stagger
		time.Sleep(1 * time.Second)
	}()
}
