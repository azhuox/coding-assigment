package account

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDailyLoadFundsChecker(t *testing.T) {
	customerID := identifier("fake-customer-id")
	date := transactionDate("2020-11-08")
	testCases := []struct {
		caseName        string
		customerAccount *customerAccount
		transaction     *loadTransaction
		err             error
	}{
		{
			caseName: "The transaction does not exceed daily load fund limit",
			customerAccount: &customerAccount{
				CustomerID: customerID,
				DailyLoadedFunds: map[transactionDate]float64{
					date: float64(0),
				},
			},
			transaction: &loadTransaction{
				ID:              "transaction-0",
				LoadAmountFloat: 555.55,
			},
			err: nil,
		},
		{
			caseName: "The transaction exceeds daily load fund limit",
			customerAccount: &customerAccount{
				CustomerID: customerID,
				DailyLoadedFunds: map[transactionDate]float64{
					date: 4000.54,
				},
			},
			transaction: &loadTransaction{
				ID:              "transaction-1",
				LoadAmountFloat: 999.47,
				currentDate:     date,
			},
			err: fmt.Errorf("exceeds maximum daily load funds ($5,000) on date %s", date.String()),
		},
	}

	checker := &dailyLoadFundsChecker{}
	for _, c := range testCases {
		err := checker.check(c.customerAccount, c.transaction)
		assert.Equal(t, c.err, err)
	}
}
