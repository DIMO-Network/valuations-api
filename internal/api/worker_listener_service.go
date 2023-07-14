package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/valuations-api/internal/core/commands"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const (
	vehicleSignalDecodingEventType = "zone.dimo.canbus.signal.update"
)

type WorkerListenerService struct {
	logger  zerolog.Logger
	handler commands.RunValuationCommandHandler
}

type VechicleSignalDecodingData struct {
	Signals []map[string]commands.RunValuationCommandRequest `json:"signals"`
}

func NewWorkerListenerService(logger zerolog.Logger, handler commands.RunValuationCommandHandler) *WorkerListenerService {
	return &WorkerListenerService{
		logger:  logger,
		handler: handler,
	}
}

func (i *WorkerListenerService) ProcessWorker(messages <-chan *message.Message) {
	for msg := range messages {
		err := i.processMessage(msg)
		if err != nil {
			i.logger.Err(err).Msg("error processing task status message")
		}
	}
}

func (i *WorkerListenerService) processMessage(msg *message.Message) error {
	// Keep the pipeline moving no matter what.
	defer func() { msg.Ack() }()

	event := new(shared.CloudEvent[VechicleSignalDecodingData])
	if err := json.Unmarshal(msg.Payload, event); err != nil {
		i.logger.Warn().Str("payload", string(msg.Payload)).Msg("failed to unmarshall processMessage payload")
		return errors.Wrap(err, "error parsing vehicle signal decoding payload")
	}

	return i.processEvent(event)
}

func (i *WorkerListenerService) processEvent(event *shared.CloudEvent[VechicleSignalDecodingData]) error {
	var (
		ctx = context.Background()
	)

	switch event.Type {
	case vehicleSignalDecodingEventType:
		command := &commands.RunValuationCommandRequest{}

		return i.handler.Execute(ctx, command)
	default:
		return fmt.Errorf("unexpected event type %s", event.Type)
	}
}
