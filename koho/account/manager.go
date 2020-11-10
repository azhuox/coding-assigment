package account

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// Manager defines the interface for managing accounts
type Manager interface {
	ProcessLoadTransactions(ctx context.Context, inputFile, outputFile string) error
}

// ManagerDefault - default account manager.
type ManagerDefault struct {
	transactionCheckers []transactionChecker
}

// NewManager - create a new instance of default account manager.
func NewManager() *ManagerDefault {
	man := &ManagerDefault{
		transactionCheckers: []transactionChecker{
			&dailyLoadFundsChecker{}, &weeklyFundsChecker{}, &dailyLoadTimeChecker{},
		},
	}

	return man
}

// ProcessLoadTransactions - process the load transactions in the given input file.
// inputFile - The file that contains load transactions that need to be processed.
// outputFile - The file that contains all the transaction results.
// error - an error that occurred during the process of transactions.
func (m *ManagerDefault) ProcessLoadTransactions(ctx context.Context, inputFile, outputFile string) error {

	// Load transactions into transaction queues
	transactionQueues, customerAccounts, totalTransactions, err := m.loadTransactionsAndCustomers(ctx, inputFile)
	if err != nil {
		return fmt.Errorf("error loading transactions: %s", err.Error())
	}

	// Prepare channel buffers for triggering multiple go routines to process load transactions.
	scheduleCh := make(chan struct{}, 50)
	transactionResultCh := make(chan *loadTransactionResult, 50)
	customersInProcess := make(map[identifier]bool, 0)
	processedCustomers := make(chan identifier, 50)
	done := make(chan struct{}, 1)

	// Open output file and trigger a go routine to process transaction results
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error openning output file %s: %s", outputFile, err)
	}
	defer func() {
		_ = outFile.Close()
	}()

	go m.processLoadTransactionsResultsRoutine(ctx, outFile, totalTransactions, transactionResultCh, processedCustomers, done)

	for totalTransactions != 0 {
		// Start at most 50 go routines to process transactions for at most customers in parallel.
		for customerID, transactionQueue := range transactionQueues {
			if !customersInProcess[customerID] && !transactionQueue.isEmpty() {
				// Get a slot from channel buffer and start a go routine to process the first transaction in the queue.
				scheduleCh <- struct{}{}
				go m.processLoadTransaction(ctx, transactionQueue.popFront(), customerAccounts[customerID], transactionResultCh, scheduleCh)
				customersInProcess[customerID] = true
				totalTransactions--
			}
		}

		// Remove processed customers from `customersInProcess` map.
		hasMoreProcessedCustomers := true
		for {
			if !hasMoreProcessedCustomers {
				break
			}
			select {
			case processedCustomer := <-processedCustomers:
				// Mark the customer "processed"
				customersInProcess[processedCustomer] = false
			default:
				hasMoreProcessedCustomers = false
			}
		}
	}

	// Wait 'processLoadTransactionsResultsRoutine' for processing all the transaction results.
	<-done

	// Reset properties and return
	return nil
}

// loadTransactionsAndCustomers - load all the transactions in the given files to the memory and
// create corresponding customers.
// Params:
//	inputFile: The file that includes some load transactions.
// Returns:
//	map[string]*transactionQueue: A map of transaction queue and each of them represents a transaction queue for a customer.
// 	map[identifier]*customerAccount: A map of customer accounts indexed by customer IDs.
//	int: Total number of transactions that needs to be processed.
//	error: Any error tha occurred during loading transactions to the memory.
func (m *ManagerDefault) loadTransactionsAndCustomers(ctx context.Context, inputFile string) (
	map[identifier]*transactionQueue, map[identifier]*customerAccount, int, error) {

	file, err := os.Open(inputFile)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("error openning file %s: %s", inputFile, err.Error())
	}
	defer func() {
		_ = file.Close()
	}()

	// Load transactions and put them into transaction queues
	transactionCount := 0
	transactionQueues := make(map[identifier]*transactionQueue, 0)
	customerAccounts := make(map[identifier]*customerAccount, 0)
	scanner := bufio.NewScanner(file)

	// Scan and load transactions
	for scanner.Scan() {
		transactionBytes := scanner.Bytes()
		transaction := loadTransaction{}
		err := json.Unmarshal(transactionBytes, &transaction)
		if err != nil {
			log.Printf("error loading transaction %s: %s", string(transactionBytes), err.Error())
			continue
		}
		if err = transaction.transformAndValidate(); err != nil {
			log.Printf("error transforming and validating transaction %#v: %s", transaction, err.Error())
		}

		// Load transactions into customer's transaction queues
		if transactionQueues[transaction.CustomerID] == nil {
			// Create the customer's transaction queue if it does not exist.
			transactionQueues[transaction.CustomerID] = newTransactionQueue(transaction.CustomerID)
		}
		transactionQueues[transaction.CustomerID].pushBack(&transaction)
		transactionCount++

		if customerAccounts[transaction.CustomerID] == nil {
			// Create customer account if it does not exist.
			customerAccounts[transaction.CustomerID] = &customerAccount{
				CustomerID:        transaction.CustomerID,
				DailyLoadedFunds:  make(map[transactionDate]float64, 0),
				WeeklyLoadedFunds: make(map[transactionDate]float64, 0),
				DailyLoadedTime:   make(map[transactionDate]uint, 0),
			}
		}
	}

	if scanner.Err() != nil {
		return nil, nil, 0, fmt.Errorf("error scanning file %s: %s", inputFile, scanner.Err().Error())
	}

	return transactionQueues, customerAccounts, transactionCount, nil
}

// processLoadTransaction - process the given transaction
// Params:
// 	transaction: The transaction that needs to be processed.
// 	customerAccount: The account of the customer who owns this transaction.
// 	transactionResultCh: A channel buffer for saving transaction results.
// 	scheduleCh: A channel buffer for controlling the number of transaction-process routines
func (m *ManagerDefault) processLoadTransaction(
	ctx context.Context, transaction *loadTransaction, customerAccount *customerAccount,
	transactionResultCh chan<- *loadTransactionResult, scheduleCh <-chan struct{}) {

	result := &loadTransactionResult{
		ID:         transaction.ID,
		CustomerID: transaction.CustomerID,
	}

	// Check whether this transaction hits some limit.
	for _, checker := range m.transactionCheckers {
		if err := checker.check(customerAccount, transaction); err != nil {
			result.Accepted = false
			result.Error = err
			goto end
		}
	}

	// Update customer account if all checks are passed.
	customerAccount.DailyLoadedFunds[transaction.currentDate] += transaction.LoadAmountFloat
	customerAccount.WeeklyLoadedFunds[transaction.mondayDateOfCurrentWeek] += transaction.LoadAmountFloat
	customerAccount.DailyLoadedTime[transaction.currentDate] += 1
	result.Accepted = true

end:
	transactionResultCh <- result

	// Release a slot
	<-scheduleCh
}

// processLoadTransactionsResultsRoutine - a routine for processing transaction results.
// Params:
//	totalTransactions: Total transactions that need to be processed.
//  transactionResultCh: A channel buffer for passing transaction results to this routine.
//  processedCustomers: A channel buffer notifying processed customers to the main routine.
//  done: A channel for notifying the main routine that all transactions results are processed.
func (m *ManagerDefault) processLoadTransactionsResultsRoutine(
	ctx context.Context, outputFile *os.File, totalTransactions int,
	transactionResultCh <-chan *loadTransactionResult, processedCustomers chan<- identifier, done chan<- struct{}) {

	enc := json.NewEncoder(outputFile)

	for totalTransactions != 0 {
		select {
		case <-ctx.Done():
			// This should never happen
			done <- struct{}{}
			return
		case result := <-transactionResultCh:
			if result.Error != nil {
				// Create an error log for failed transaction.
				log.Printf("error processing transaction %s for customer %s: %s",
					result.ID.String(), result.CustomerID.String(), result.Error.Error())
			}

			if err := enc.Encode(result); err != nil {
				log.Printf("error writing transaction %s to the output file for customer %s: %s",
					result.ID.String(), result.CustomerID.String(), result.Error.Error())
			}

			processedCustomers <- result.CustomerID
			totalTransactions--
		}
	}

	done <- struct{}{}
}
