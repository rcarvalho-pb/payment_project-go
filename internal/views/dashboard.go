package views

import (
	"fmt"

	"github.com/rcarvalho-pb/payment_project-go/internal/domain/invoice"
)

type DashboardSummary struct {
	InvoiceCount   int
	TotalAmount    string
	OpenCount      int
	PaidCount      int
	FailureCount   int
	ProcessingRate string
}

func BuildDashboardSummary(invoices []*invoice.Invoice) DashboardSummary {
	total := int64(0)
	openCount := 0
	paidCount := 0
	failureCount := 0

	for _, inv := range invoices {
		total += inv.Amount

		switch inv.Status {
		case invoice.StatusPaid:
			paidCount++
		case invoice.StatusFailed, invoice.StatusCanceled:
			failureCount++
		case invoice.StatusPending, invoice.StatusProcessing:
			openCount++
		}
	}

	rate := "0%"
	if len(invoices) > 0 {
		rate = fmt.Sprintf("%d%%", (paidCount*100)/len(invoices))
	}

	return DashboardSummary{
		InvoiceCount:   len(invoices),
		TotalAmount:    fmt.Sprintf("R$ %d", total),
		OpenCount:      openCount,
		PaidCount:      paidCount,
		FailureCount:   failureCount,
		ProcessingRate: rate,
	}
}
