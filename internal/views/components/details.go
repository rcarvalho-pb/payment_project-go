package components

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/rcarvalho-pb/payment_project-go/internal/domain/invoice"
	"github.com/rcarvalho-pb/payment_project-go/internal/domain/payment"
)

type InvoiceDetailsData struct {
	Invoice  *invoice.Invoice
	Payments []*payment.Payment
}

func (d InvoiceDetailsData) AttemptCount() int {
	return len(d.Payments)
}

func (d InvoiceDetailsData) LastPaymentStatus() string {
	if len(d.Payments) == 0 {
		return "Aguardando primeira tentativa"
	}

	return d.Payments[len(d.Payments)-1].Status.String()
}

func (d InvoiceDetailsData) LastAttemptNumber() int {
	if len(d.Payments) == 0 {
		return 0
	}

	return d.Payments[len(d.Payments)-1].Attempt
}

func (d InvoiceDetailsData) RetryCount() int {
	attempts := d.AttemptCount()
	if attempts == 0 {
		return 0
	}

	return attempts - 1
}

func SortPaymentsByAttempt(payments []*payment.Payment) []*payment.Payment {
	cloned := append([]*payment.Payment(nil), payments...)
	sort.Slice(cloned, func(i, j int) bool {
		if cloned[i].Attempt == cloned[j].Attempt {
			return cloned[i].CreatedAt.Before(cloned[j].CreatedAt)
		}

		return cloned[i].Attempt < cloned[j].Attempt
	})

	return cloned
}

func formatDateTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}

	return t.Format("02/01/2006 15:04:05")
}

func formatMoneyBRL(value int64) string {
	return fmt.Sprintf("R$ %d", value)
}

func paymentStatusClass(status string) string {
	switch strings.ToUpper(status) {
	case "SUCCESS":
		return "bg-success"
	case "PROCESSING":
		return "bg-warning text-dark"
	case "TEMPORARY_FAILED":
		return "bg-warning text-dark"
	case "FAILED":
		return "bg-danger"
	default:
		return "bg-secondary"
	}
}
