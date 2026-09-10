package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Srivastava-samarth/sampay/dto"
)

type ReportService struct {}

func NewReportService () *ReportService{
	return &ReportService{}
}



func (rs *ReportService) GenerateReconReport(
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

func (rs *ReportService) GenerateReconEmailBody(
    report *dto.ReconReport,
) (string, error) {
    if report == nil {
        return "", errors.New("recon report cannot be nil")
    }

    var builder strings.Builder

    builder.WriteString("SAMPay Reconciliation Report\n")
    builder.WriteString("============================\n\n")

    builder.WriteString(fmt.Sprintf(
        "Recon Date: %s\n",
        report.ReconDate,
    ))

    builder.WriteString(fmt.Sprintf(
        "Total Transactions: %d\n",
        report.TotalTransactions,
    ))

    builder.WriteString(fmt.Sprintf(
        "Matched: %d\n",
        report.MatchedCount,
    ))

    builder.WriteString(fmt.Sprintf(
        "Failed: %d\n",
        report.FailedCount,
    ))

    builder.WriteString(fmt.Sprintf(
        "Total Debit: %s\n",
        report.TotalDebit,
    ))

    builder.WriteString(fmt.Sprintf(
        "Total Credit: %s\n",
        report.TotalCredit,
    ))

    builder.WriteString(fmt.Sprintf(
        "Difference: %s\n",
        report.Difference,
    ))

    builder.WriteString(fmt.Sprintf(
        "Status: %s\n\n",
        report.Status,
    ))

    if report.FailedCount > 0 {
        builder.WriteString("Failed Transactions\n")
        builder.WriteString("-------------------\n")

        for _, failure := range report.FailedTransactions {
            builder.WriteString(fmt.Sprintf(
                "Ledger Transaction ID: %s\n",
                failure.LedgerTransactionID,
            ))

            builder.WriteString(fmt.Sprintf(
                "Reference ID: %s\n",
                failure.ReferenceID,
            ))

            builder.WriteString(fmt.Sprintf(
                "Debit: %s\n",
                failure.TotalDebit,
            ))

            builder.WriteString(fmt.Sprintf(
                "Credit: %s\n",
                failure.TotalCredit,
            ))

            builder.WriteString(fmt.Sprintf(
                "Difference: %s\n\n",
                failure.Difference,
            ))
        }
    }

    return builder.String(), nil
}
