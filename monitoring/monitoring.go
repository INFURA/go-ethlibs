package monitoring

import (
	"fmt"
	"log"

	"github.com/ConsenSys/go-ethlibs/conf"
	"github.com/DataDog/datadog-go/statsd"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type Monitoring struct {
	DatadogClient *statsd.Client
	CheckName     string
}

func New(config conf.MonitoringEnvSpec, checkName string) (*statsd.Client, func(), error) {
	datadogClient, err := statsd.New("") // uses env vars
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to init dd statsd: %w", err)
	}

	tracer.Start(
		tracer.WithEnv(config.Env),
		tracer.WithService(checkName),
	)

	return datadogClient, func() {
		datadogClient.Close()
		tracer.Stop()
	}, nil
}

func (m *Monitoring) PostCheckSuccess(config conf.MonitoringEnvSpec, test string) {
	log.Printf("SUCCESS(%s, %v)\n", test, m.CheckName)

	checkTags := []string{
		fmt.Sprintf("env:%s", config.Env),
		fmt.Sprintf("network:%s", config.Network),
		fmt.Sprintf("test:%s", test),
	}

	// nolint:exhaustivestruct // statsd will fill the rest of the fields
	err := m.DatadogClient.ServiceCheck(&statsd.ServiceCheck{
		Name:     m.CheckName,
		Status:   statsd.Ok,
		Tags:     checkTags,
		Hostname: "k8s",
	})
	if err != nil {
		log.Panic(err.Error())
	}
}

func (m *Monitoring) PostCheckFailure(config conf.MonitoringEnvSpec, test string, reasonErr error) {
	reason := reasonErr.Error()

	log.Printf("FAILURE(%s): %s\n", test, reason)

	checkTags := []string{
		fmt.Sprintf("env:%s", config.Env),
		fmt.Sprintf("network:%s", config.Network),
		fmt.Sprintf("test:%s", test),
	}

	// nolint:exhaustivestruct // statsd will fill the rest of the fields
	err := m.DatadogClient.ServiceCheck(&statsd.ServiceCheck{
		Name:     m.CheckName,
		Status:   statsd.Critical,
		Tags:     checkTags,
		Message:  reason,
		Hostname: "k8s",
	})
	if err != nil {
		log.Panic(err.Error())
	}
}
