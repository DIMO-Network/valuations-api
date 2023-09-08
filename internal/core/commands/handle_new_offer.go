package commands

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/DIMO-Network/valuations-api/internal/infrastructure/db/models"
	"github.com/nats-io/nats.go"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type SaveOfferPayload struct {
	UserDeviceID string `json:"userDeviceId"`
	offerData    map[string]interface{}
}

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
			var payload SaveOfferPayload
			err := json.Unmarshal(msg.Data, &payload)
			if err != nil {
				return err
			}

			err = h.handleSaveOffer(ctx, payload)
			if err != nil {
				return err
			}

			if err := msg.Ack(); err != nil {
				h.logger.Err(err).Msg("message ack failed")
			}
		}
	}
}

func (h *runValuationCommandHandler) handleSaveOffer(ctx context.Context, payload SaveOfferPayload) error {

	userDevice, err := h.userDeviceService.GetUserDevice(ctx, payload.UserDeviceID)

	if err != nil {
		return err
	}

	existingPricingData, _ := models.Valuations(
		models.ValuationWhere.Vin.EQ(*userDevice.Vin),
		models.ValuationWhere.DrivlyPricingMetadata.IsNotNull(),
		qm.OrderBy("updated_at desc"), qm.Limit(1)).
		One(context.Background(), h.DBS().Writer)

	if existingPricingData == nil {
		return errors.New("no pricing data found for this user device")
	}

	// update existing pricing data
	_ = existingPricingData.OfferMetadata.Marshal(payload.offerData)

	_, err = existingPricingData.Update(ctx, h.DBS().Writer, boil.Infer())

	if err != nil {
		return err
	}
	return nil

}
