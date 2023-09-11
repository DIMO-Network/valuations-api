package commands

import (
	"context"
	"encoding/json"

	core "github.com/DIMO-Network/valuations-api/internal/core/models"
	"github.com/nats-io/nats.go"
)

func (h *runValuationCommandHandler) ExecuteOfferSync(ctx context.Context) error {
	sub, err := h.NATSSvc.JetStream.PullSubscribe(h.NATSSvc.OfferSubject, h.NATSSvc.OfferDurableConsumer, nats.AckWait(h.NATSSvc.AckTimeout))
	if err != nil {
		return err
	}

	for {
		msgs, err := sub.Fetch(10, nats.MaxWait(h.NATSSvc.AckTimeout))
		if err != nil {
			return err
		}

		for _, msg := range msgs {
			var payload core.OfferRequest
			err := json.Unmarshal(msg.Data, &payload)
			if err != nil {
				return err
			}

			status, err := h.drivlyValuationService.PullOffer(ctx, payload.VIN)

			if err != nil {
				h.nak(msg)
				h.logger.Err(err).Str("payload", string(msg.Data)).Msg("failed to process offer request")
				return err
			}

			h.logger.Info().Str("payload", string(msg.Data)).Msgf("processing offer request with status %s", status)

			if err := msg.Ack(); err != nil {
				h.logger.Err(err).Msg("message ack failed")
			}
		}
	}
}
