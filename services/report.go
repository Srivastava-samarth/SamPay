package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Srivastava-samarth/sampay/dto"
)

func (rs *Services) GenerateReconReport(
	result *dto.ReconResult,
) (*dto.ReconReport, error) {
	if result == nil {
		return nil, errors.New("recon result cannot be nil")
	}

	status := "PASSED"

	if *result.FailedCount > 0 {
		status = "FAILED"
	}

	difference := result.TotalDebit.Sub(result.TotalCredit)

	TotalTransaction := result.TotalTransactions
	MatchedCount := result.MatchedCount
	FailedCount := result.FailedCount
	return &dto.ReconReport{
		ReconDate:          result.ReconDate.Format("2006-01-02"),
		TotalTransactions:  *TotalTransaction,
		MatchedCount:       *MatchedCount,
		FailedCount:        *FailedCount,
		TotalDebit:         result.TotalDebit.String(),
		TotalCredit:        result.TotalCredit.String(),
		Difference:         difference.String(),
		Status:             status,
		FailedTransactions: result.FailedTransactions,
	}, nil
}

func (*Services) GenerateReconEmailBody(
	report *dto.ReconReport,
) (string, error) {
	if report == nil {
		return "", errors.New("recon report cannot be nil")
	}

	var builder strings.Builder

	builder.WriteString("SAMPay Reconciliation Report\n")
	builder.WriteString("============================\n\n")

	fmt.Fprintf(&builder, "Recon Date: %s\n", report.ReconDate)
	fmt.Fprintf(&builder, "Total Transactions: %d\n", report.TotalTransactions)
	fmt.Fprintf(&builder, "Matched: %d\n", report.MatchedCount)
	fmt.Fprintf(&builder, "Failed: %d\n", report.FailedCount)
	fmt.Fprintf(&builder, "Total Debit: %s\n", report.TotalDebit)
	fmt.Fprintf(&builder, "Total Credit: %s\n", report.TotalCredit)
	fmt.Fprintf(&builder, "Difference: %s\n", report.Difference)
	fmt.Fprintf(&builder, "Status: %s\n\n", report.Status)

	if report.FailedCount > 0 {
		builder.WriteString("Failed Transactions\n")
		builder.WriteString("-------------------\n")

		for _, failure := range report.FailedTransactions {
			fmt.Fprintf(
				&builder,
				"Ledger Transaction ID: %s\n",
				failure.LedgerTransactionID,
			)
			fmt.Fprintf(
				&builder,
				"Reference ID: %s\n",
				failure.ReferenceID,
			)
			fmt.Fprintf(
				&builder,
				"Debit: %s\n",
				failure.TotalDebit,
			)
			fmt.Fprintf(
				&builder,
				"Credit: %s\n",
				failure.TotalCredit,
			)
			fmt.Fprintf(
				&builder,
				"Difference: %s\n\n",
				failure.Difference,
			)
		}
	}

	result := builder.String()
	return result, nil
}
