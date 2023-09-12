package services

import (
	"github.com/pkg/errors"
	"time"

	"github.com/DIMO-Network/valuations-api/internal/config"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
)

type NATSService struct {
	log                      *zerolog.Logger
	JetStream                nats.JetStreamContext
	JetStreamName            string
	ValuationSubject         string
	OfferSubject             string
	AckTimeout               time.Duration
	ValuationDurableConsumer string
	OfferDurableConsumer     string
}

func NewNATSService(settings *config.Settings, log *zerolog.Logger) (*NATSService, error) {
	n, err := nats.Connect(settings.NATSURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to "+settings.NATSURL)
	}

	js, err := n.JetStream()
	if err != nil {
		return nil, err
	}

	_, err = js.AddStream(&nats.StreamConfig{
		Name:      settings.NATSStreamName,
		Retention: nats.WorkQueuePolicy,
		Subjects:  []string{settings.NATSValuationSubject, settings.NATSOfferSubject},
	})

	if err != nil {

		if errors.Is(err, nats.ErrStreamNameAlreadyInUse) {

			_, err = js.UpdateStream(&nats.StreamConfig{
				Name:      settings.NATSStreamName,
				Retention: nats.WorkQueuePolicy,
				Subjects:  []string{settings.NATSValuationSubject, settings.NATSOfferSubject},
			})

			if err != nil {
				return nil, err
			}
		}

		return nil, err
	}

	to, err := time.ParseDuration(settings.NATSAckTimeout)
	if err != nil {
		return nil, err
	}

	natsSvc := &NATSService{
		log:                      log,
		JetStream:                js,
		JetStreamName:            settings.NATSStreamName,
		ValuationSubject:         settings.NATSValuationSubject,
		OfferSubject:             settings.NATSOfferSubject,
		AckTimeout:               to,
		ValuationDurableConsumer: settings.NATSValuationDurableConsumer,
		OfferDurableConsumer:     settings.NATSOfferDurableConsumer,
	}

	return natsSvc, nil
}
