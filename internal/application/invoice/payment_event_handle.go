package invoice

import (
	"errors"
	"log"

	"github.com/rcarvalho-pb/payment_project-go/internal/domain/event"
	"github.com/rcarvalho-pb/payment_project-go/internal/domain/invoice"
)

type PaymentEventHandler struct {
	Repo invoice.Repository
}

func (h *PaymentEventHandler) Handle(evt *event.Event) error {
	switch evt.Type {
	case event.PaymentSucceeded:
		log.Println("payment event handler: succeeded payment")
		payment, ok := evt.Payload.(*event.PaymentSucceededPayload)
		if !ok {
			log.Println("payment event handler: failed to cast payload to succeeded")
			return errors.New("invalid payload for PaymentSucceeded")
		}

		err := h.Repo.UpdateStatus(payment.InvoiceID, invoice.StatusPaid)
		if err != nil {
			log.Println("error saving in db succeeded")
		}
		return err
	case event.PaymentFailed:
		log.Println("payment event handler: failed payment")
		payment, ok := evt.Payload.(*event.PaymentFailedPayload)
		if !ok {
			log.Println("payment event handler: failed to cast payload to failed")
			return errors.New("invalid payload for PaymentFailed")
		}

		if !payment.Retryable {
			err := h.Repo.UpdateStatus(payment.InvoiceID, invoice.StatusFailed)
			if err != nil {
				log.Println("error saving in db failed")
			}
			return err
		}
		return nil
	}
	return nil
}
