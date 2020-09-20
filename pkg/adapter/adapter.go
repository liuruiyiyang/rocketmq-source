/*
Copyright 2019 The Knative Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
		http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package adapter implements a sample receive adapter that generates events
// at a regular interval.
package adapter

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"go.uber.org/zap"
	"knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"
)

type envConfig struct {
	// Include the standard adapter.EnvConfig used by all adapters.
	adapter.EnvConfig

	// Interval between events, for example "5s", "100ms"
	Interval time.Duration `envconfig:"INTERVAL" required:"true"`
}

func NewEnv() adapter.EnvConfigAccessor { return &envConfig{} }

// Adapter generates events at a regular interval.
type Adapter struct {
	client   cloudevents.Client
	interval time.Duration
	logger   *zap.SugaredLogger

	sequence int64
}

func (a *Adapter) newEvent(message *primitive.MessageExt) cloudevents.Event {
	event := cloudevents.NewEvent()
	event.SetID(message.MsgId)
	event.SetType("dev.knative.rocketmq")
	event.SetSource("rocketmq.knative.dev/topic-source")

	if err := event.SetData(cloudevents.ApplicationJSON, message); err != nil {
		a.logger.Errorw("failed to set data")
	}
	a.sequence++
	return event
}

// Start runs the adapter.
// Returns if ctx is cancelled or Send() returns an error.
func (a *Adapter) Start(ctx context.Context) error {
	a.logger.Infow("Starting to consume message from RocketMQ and pass to CloudEvent...")
	a.logger.Infow("Configs: ", zap.String("interval", a.interval.String()))

	c, _ := rocketmq.NewPushConsumer(
		consumer.WithGroupName("knative_source"),
		consumer.WithNsResovler(primitive.NewPassthroughResolver([]string{"30.25.106.54:9876"})),
		// consumer.WithSuspendCurrentQueueTimeMillis(10000),
	)
	err := c.Subscribe("newOne", consumer.MessageSelector{}, func(ctx context.Context,
		msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for i := range msgs {
			event := a.newEvent(msgs[i])
			a.logger.Infow("Sending new event", zap.String("event", event.String()))
			if result := a.client.Send(context.Background(), event); !cloudevents.IsACK(result) {
				a.logger.Infow("failed to send event", zap.String("event", event.String()), zap.Error(result))
				// We got an error but it could be transient, try again next interval.
				continue
			}
		}
		return consumer.ConsumeSuccess, nil
	})
	if err != nil {
		fmt.Println(err.Error())
	}
	// Note: start after subscribe
	err = c.Start()

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	for true {
		time.Sleep(time.Hour)
	}
	a.logger.Info("Shutting down rocketmq push consumer...")
	err = c.Shutdown()
	if err != nil {
		fmt.Printf("shutdown Consumer error: %s", err.Error())
	}
	a.logger.Info("rocketmq push consumer closed")
	return err
}

func NewAdapter(ctx context.Context, aEnv adapter.EnvConfigAccessor, ceClient cloudevents.Client) adapter.Adapter {
	env := aEnv.(*envConfig) // Will always be our own envConfig type
	logger := logging.FromContext(ctx)
	logger.Infow("RocketMQ Source")
	return &Adapter{
		interval: env.Interval,
		client:   ceClient,
		logger:   logger,
	}
}
