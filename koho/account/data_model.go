package account

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type identifier string

// String - convert the identifier to string.
func (c identifier) String() string {
	return string(c)
}

// customerAccount - Customer's customerAccount
type customerAccount struct {
	CustomerID        identifier
	DailyLoadedFunds  map[transactionDate]float64
	WeeklyLoadedFunds map[transactionDate]float64
	DailyLoadedTime   map[transactionDate]uint
}

/****************************************************************************************/

// loadTransaction - a transaction to load funds
//
type loadTransaction struct {
	ID                      identifier `json:"id"`
	CustomerID              identifier `json:"customer_id"`
	LoadAmount              string     `json:"load_amount"`
	LoadAmountFloat         float64
	Time                    time.Time `json:"time"`
	currentDate             transactionDate
	mondayDateOfCurrentWeek transactionDate
}

func (t *loadTransaction) transformAndValidate() error {
	if t.ID == "" {
		return fmt.Errorf("transaction's ID is empty")
	}
	if t.CustomerID == "" {
		return fmt.Errorf("transaction's customer ID is empty")
	}
	if t.LoadAmount == "" {
		return fmt.Errorf("transaction's load amount is empty")
	}
	if t.Time.IsZero() {
		return fmt.Errorf("transaction's time is empty")
	}

	// Convert money from string format to a float.
	var err error
	t.LoadAmountFloat, err = strconv.ParseFloat(strings.TrimSpace(t.LoadAmount[1:]), 64)
	if err != nil {
		return fmt.Errorf("transaction's load amount is not a valid number")
	}
	t.currentDate = dateFromTime(t.Time)
	t.mondayDateOfCurrentWeek = mondayDateFromTime(t.Time)

	return nil
}

// Helper functions for `loadTransaction`

type transactionDate string

// String - convert the transaction date to string.
func (t transactionDate) String() string {
	return string(t)
}

// dateFromTime - return date formatted with "yyyy-mm-dd" for the given time.
func dateFromTime(t time.Time) transactionDate {
	return transactionDate(fmt.Sprintf("%d-%d-%d", t.Year(), t.Month(), t.Day()))
}

// weekStartDateFromTime - return date of monday of current week based on the given time.
func mondayDateFromTime(t time.Time) transactionDate {
	var monday time.Time
	if t.Weekday() == time.Sunday {
		monday = t.AddDate(0, 0, -6)
	} else {
		monday = t.AddDate(0, 0, int(time.Monday-t.Weekday()))
	}

	return transactionDate(fmt.Sprintf("%d-%d-%d", monday.Year(), monday.Month(), monday.Day()))
}

/****************************************************************************************/

// transaction checker
type transactionChecker interface {
	check(a *customerAccount, t *loadTransaction) error
}

// dailyLoadFundsChecker - check whether given transaction hit daily load fund limit.
type dailyLoadFundsChecker struct{}

func (c *dailyLoadFundsChecker) check(a *customerAccount, t *loadTransaction) error {
	if (a.DailyLoadedFunds[t.currentDate] + t.LoadAmountFloat) > float64(5000) {
		return fmt.Errorf("exceeds maximum daily load funds ($5,000) on date %s", t.currentDate.String())
	}
	return nil
}

// weeklyFundsChecker - check whether given transaction hit weekly load fund limit.
type weeklyFundsChecker struct{}

func (c *weeklyFundsChecker) check(a *customerAccount, t *loadTransaction) error {
	if (a.WeeklyLoadedFunds[t.mondayDateOfCurrentWeek] + t.LoadAmountFloat) > float64(20000) {
		return fmt.Errorf("exceeds maximum daily load funds ($20,000) on week which monday is %s",
			t.mondayDateOfCurrentWeek.String())
	}
	return nil
}

// dailyLoadTimeChecker - check whether given transaction hit daily load time limit.
type dailyLoadTimeChecker struct{}

func (c *dailyLoadTimeChecker) check(a *customerAccount, t *loadTransaction) error {

	if (a.DailyLoadedTime[t.currentDate] + 1) > 3 {
		return fmt.Errorf("exceeds maximum daily load time (3) on date %s", t.currentDate.String())
	}
	return nil
}

type loadTransactionResult struct {
	ID         identifier `json:"id"`
	CustomerID identifier `json:"customer_id"`
	Accepted   bool       `json:"accepted"`
	Error      error      `json:"-"`
}
