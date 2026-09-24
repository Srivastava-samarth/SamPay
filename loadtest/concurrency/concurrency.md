# Financial Concurrency Load Test

## Objective

Verify that concurrent financial operations do not cause incorrect wallet or bank-account balances.

## Test

* Send multiple payment requests concurrently against the same wallet.
* Run the test with increasing concurrency levels.
* Verify that balance updates remain consistent.
* Verify that no amount is lost, duplicated, or incorrectly reserved.

## Protection

Financial balance operations use PostgreSQL row-level locking (`FOR UPDATE`) inside the same database transaction to prevent concurrent operations from reading and modifying the same balance simultaneously.

## Expected Behavior

* Each valid request is processed exactly once.
* Wallet balances remain financially consistent.
* Concurrent requests cannot overspend the available balance.
* Reserved and available balances remain consistent.
* Database transactions maintain atomicity when an operation fails.

## Result

The concurrency test passed after introducing row-level database locking.

Concurrent payment requests were executed successfully without balance inconsistencies or double spending.

## Conclusion

The financial transaction flow is protected against concurrent balance updates and can safely handle concurrent payment operations at the tested concurrency levels.
