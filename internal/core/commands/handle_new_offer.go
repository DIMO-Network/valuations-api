package commands

import (
	"context"
	"errors"

	"encoding/json"

	core "github.com/DIMO-Network/valuations-api/internal/core/models"

	"github.com/nats-io/nats.go"
)

// TODO deprecated code, remove once new stuff working

func (h *runValuationCommandHandler) ExecuteOfferSync(ctx context.Context) error {
	sub, err := h.NATSSvc.JetStream.PullSubscribe(h.NATSSvc.OfferSubject, h.NATSSvc.OfferDurableConsumer,
		nats.AckWait(h.NATSSvc.AckTimeout))
	// nats.MaxDeliver(2) if add this get error: configuration requests max deliver to be 2, but consumer's value is -1 .... but where is the consumer

	if err != nil {
		h.logger.Err(err).Msg("failed to subscribe to nats at offer sync")
		return err
	}

	h.logger.Info().Msg("subscribed to nats at offer sync")

	for {
		msgs, err := sub.Fetch(1, nats.MaxWait(h.NATSSvc.AckTimeout))

		if err != nil {
			if errors.Is(err, nats.ErrTimeout) {
				h.logger.Debug().Msg("no messages found at offer sync")
				continue
			}

			h.logger.Err(err).Msg("failed to fetch messages from nats at offer sync")
			return err
		}

		for _, msg := range msgs {
			mtd, err := msg.Metadata()

			if err != nil {
				ackErr := msg.Ack() // if we nak, then it just keeps retrying forever, and it isn't viable to set the MaxDeliver
				if ackErr != nil {
					h.logger.Err(err).Msg("message ack failed")
				}
				h.logger.Err(err).Msg("failed to process offer request due to invalid payload")
				continue
			}

			select {
			case <-ctx.Done():
				h.logger.Info().Msg("context cancelled at offer sync")
				return nil
			default:
				var payload core.OfferRequest

				err := json.Unmarshal(msg.Data, &payload)

				if err != nil {
					_ = msg.Ack()
					h.logger.Err(err).Str("payload", string(msg.Data)).Msg("failed to process offer request due to invalid payload - not redelivering")
					continue
				}

				status, err := h.drivlyValuationService.PullOffer(ctx, payload.UserDeviceID, 0, "")

				if err != nil && status != core.SkippedDataPullStatus {
					// todo: improvement - depending on status returned decide whether to redeliver using h.nak(msg)
					_ = msg.Ack()
					h.logger.Err(err).Str("payload", string(msg.Data)).Msg("failed to process offer request due to internal error - not redelivering")
					continue
				}

				h.logger.Info().Str("payload", string(msg.Data)).Msgf("processing offer request %v with status %s", mtd.NumDelivered, status)

				if err := msg.Ack(); err != nil {
					h.logger.Err(err).Msg("message ack failed for offer sync")
				}
			}
		}
	}
}
