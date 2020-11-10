# Velocity Limits Checker

This folder contains a velocity limits checker program that accepts or declines attempts to load funds into customers' accounts in real-time.

## How to Run It.

Cd to the current directory and run command `go run main.go -input_file <input_file_path>`. 
It will process transactions in the given file and write the results to [output.txt](./output.txt) file.

## Unit Tests

I did not write enough unit tests to cover to all the code because of time limitation. 
Instead, I wrote a unit test in [data_model_test.go](./account/data_model_test.go) to demonstrate that 
I use [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests) method to write unit tests.
This method puts all the test cases in a table (list) and loop through them to run every test case. The major benefit
of this method is it groups all the test cases into a test function and it is very easy to setup these test cases with this method.

## Functional Tests

I tested the following use cases:

1. A maximum of $5,000 can be loaded per day
2. A maximum of $20,000 can be loaded per week
3. A maximum of 3 loads can be performed per day, regardless of amount

If you run the program with the transactions in [input.txt](./input.txt), you will find the following errors in the error log:

```
error processing transaction 9307 for customer 528: exceeds maximum daily load funds ($5,000) on date 2000-2-8
error processing transaction 29260 for customer 777: exceeds maximum daily load funds ($20,000) on week which monday is 2000-10-9
error processing transaction 29261 for customer 777: exceeds maximum daily load funds ($20,000) on week which monday is 2000-10-9
error processing transaction 29262 for customer 777: exceeds maximum daily load funds ($20,000) on week which monday is 2000-10-9
error processing transaction 29269 for customer 888: exceeds maximum daily load time (3) on date 2000-10-12
```

If you check the transactions in the input file, you will find:

- customer `528` did violate the first rule on transaction `9307`.
- customer `777` did violate the second rule on transaction `29260`, `29261` and `29262`
- customer `888` did violate the third rule on transaction `29269` 

These results prove that my program cover all the user cases.

## How My Program Works

My program works in the following way:

1. Open the input file and load all the transactions and customer accounts into two Golang Maps.
The first map represent transaction queues and each of them represents a customer's transaction queue.
All transactions of a customer will be loaded to this customer's transaction queue and will be processed in sequence based on transaction time. 
The second map represents customer accounts and each of them represents a customer's account. 
This account has the statics of daily and weekly fund limits.

2. Main routine uses go routines and channel buffer to trigger at most 50 goroutines to process at most 50 customer' transactions. 
It ensures that a customer can only have at most one transaction that is being processed at any time. This provides the guarantee that
all the transactions of a customer are processed in sequence based on transaction time.

3. A routine is scheduled to process transaction results. For each transaction result, 
it will log out the error if this transaction result has an error and write the result to the output file.

4. The `transactionChecker` interface is defined for realizing all kinds of velocity limits checker. 
With this interface, we can easily add & remove any checker we want without modifying any code in `account.Manager`.    

