# Payment Workflow Reliability Testing

## 1. Objective

Validate the reliability and consistency of the SamPay payment workflow under normal load, concurrent execution, idempotent requests, activity failures, and Temporal service interruptions.

The primary focus is to verify that:

* Payment workflows complete correctly under load.
* Concurrent payments do not cause balance inconsistencies.
* Duplicate requests do not create duplicate payments.
* Temporal activity retries work as expected.
* Temporal service interruptions do not result in duplicate financial transactions or stuck payments.

---

## 2. Payment Workflow

The payment workflow currently executes the following activities:

```text
ValidatePaymentRequest
        ↓
CalculateFees
        ↓
CreatePayment
        ↓
ExecutePayment
        ↓
UpdatePaymentStatus (on execution failure)
```

`ExecutePayment` performs the financial operations inside a single database transaction.

---

## 3. Test Environment

| Component         | Configuration                        |
| ----------------- | ------------------------------------ |
| Application       | SamPay                               |
| Language          | Go                                   |
| Workflow Engine   | Temporal                             |
| Database          | PostgreSQL                           |
| Payment API       | `POST /api/v1/{merchant_id}/payment` |
| Temporal Server   | Local Temporal Development Server    |
| Payment Load Test | Custom Go load-test client           |

---

# 4. Test Cases

## 4.1 Sustained Payment Load

### Objective

Verify that the payment workflow can process a sustained stream of payment requests without errors under normal operating conditions.

### Test

| Parameter |      Value |
| --------- | ---------: |
| Rate      |      1 RPS |
| Duration  |  5 seconds |
| Rate      |      1 RPS |
| Duration  | 30 seconds |
| Rate      |      2 RPS |
| Duration  | 30 seconds |

### Result

**PASS**

All requests completed successfully during the sustained-load tests.

Observed average latency remained in the approximate range of **60–80 ms** during the successful load tests.

---

## 4.2 Concurrent Payments

### Objective

Verify that simultaneous payment requests cannot cause wallet balance races or inconsistent financial state.

### Test

10 payment requests were executed concurrently against the same wallet.

### Initial Issue

Concurrent execution initially exposed a balance race condition.

Multiple requests could read the same available balance before updating it.

### Fix

Row-level PostgreSQL locks were introduced using `FOR UPDATE` when retrieving financial records.

Wallet and bank-account updates are now performed within the same database transaction.

### Result

**PASS**

10 concurrent payments of ₹2,000 completed successfully without an observed balance inconsistency.

---

## 4.3 Concurrent Auto-Top-Up

### Objective

Verify that concurrent payments requiring wallet auto-top-up do not result in an inconsistent wallet or bank balance.

### Test

Concurrent payment requests were executed while the wallet balance was insufficient, causing the auto-top-up mechanism to be triggered.

### Result

**PASS**

The concurrency scenario completed successfully after the row-level locking changes.

---

## 4.4 Idempotency

### Objective

Verify that multiple requests using the same idempotency key do not create multiple payments.

### Test

10 concurrent requests were sent using the same idempotency key.

### Observed Result

```text
1 request  → 200
8 requests → 409 REQUEST_IN_PROGRESS
1 request  → 200 with the completed payment response
```

Only **one payment** was created.

### Result

**PASS**

The idempotency mechanism correctly prevents duplicate payment creation.

The current behavior intentionally returns `409 REQUEST_IN_PROGRESS` when another request is already processing the same idempotency key.

---

## 4.5 Temporal Activity Failure & Retry

### Objective

Verify that Temporal automatically retries a failed payment activity according to the configured retry policy.

### Retry Configuration

```text
Initial Interval:     2 seconds
Backoff Coefficient:  2.0
Maximum Interval:     30 seconds
Maximum Attempts:     5
```

### Test

A temporary failure was injected into the `ExecutePayment` activity.

### Observed Result

```text
ExecutePayment Attempt 1
        ↓
temporary failure
        ↓ ~2 seconds
ExecutePayment Attempt 2
        ↓
success
```

The Temporal dashboard showed the first attempt as failed and the second attempt as successful.

The API ultimately returned HTTP `200`.

### Result

**PASS**

Temporal activity retry behavior works according to the configured retry policy.

### Limitation

This test verifies Temporal's retry mechanism but does not test failure after financial side effects have already been committed.

---

# 4.6 Temporal Server Outage & Recovery

### Objective

Verify application behavior when the Temporal server becomes unavailable while payment workflows are actively running.

### Test

A financial load test was executed at:

```text
2 RPS for 30 seconds
```

Temporal was stopped while workflows were actively being processed and subsequently restarted.

### Load Test Result

| Metric                      |  Result |
| --------------------------- | ------: |
| Total Requests              |      60 |
| HTTP 200                    |      23 |
| HTTP 500                    |      37 |
| Completed Payments          |      24 |
| Pending Payments            |       0 |
| Failed Payments             |       0 |
| Payment Ledger Transactions |      24 |
| Average Latency             |  9.46 s |
| Maximum Latency             | 27.72 s |

### Observed Behavior

While Temporal was unavailable, affected API requests eventually returned HTTP `500`, with many requests taking approximately 10 seconds before failing.

After Temporal was restarted, workflows were able to recover and complete.

No payment remained in a `PENDING` state.

The number of completed payments matched the number of payment ledger transactions:

```text
Completed Payments        = 24
Payment Ledger Transactions = 24
```

### Result

**PASS — with an API behavior observation**

The test did not reveal duplicate payment ledger transactions or stuck payments.

### Important Observation

An HTTP `500` response does not necessarily mean that the underlying payment workflow failed.

A possible sequence is:

```text
Client
  ↓
Payment API
  ↓
Temporal Workflow
  ↓
Temporal unavailable
  ↓
API timeout → HTTP 500
  ↓
Temporal recovers
  ↓
Workflow continues
  ↓
Payment completes
```

Therefore, a client may receive an HTTP `500` while the underlying payment eventually reaches `COMPLETED`.

This makes the idempotency mechanism particularly important for clients retrying payments after ambiguous failures.

---

# 5. Current Test Summary

| Test                              | Result |
| --------------------------------- | ------ |
| Sustained Payment Load            | PASS   |
| Concurrent Payments               | PASS   |
| Concurrent Auto-Top-Up            | PASS   |
| Idempotency                       | PASS   |
| Temporal Activity Failure & Retry | PASS   |
| Temporal Server Outage & Recovery | PASS*  |

`*` No duplicate or stuck financial state was observed, but the HTTP/API behavior during Temporal outages should be considered separately from the final workflow state.

---

# 6. Remaining Reliability Test

The major remaining scenario is **failure during or immediately around financial execution**.

The purpose is to verify that a Temporal activity remains safe if execution fails around the point where financial side effects occur.

This should not be implemented by simply inserting a failure after committing financial changes. The test must first establish how duplicate execution is prevented.

The key property to validate is:

```text
Activity execution
      ↓
Financial transaction
      ↓
Failure / retry
      ↓
Retry
      ↓
No duplicate financial effect
```

This test should be designed separately before modifying `ExecutePayment`.

---

# 7. Acceptance Criteria

The payment workflow should satisfy the following properties:

* Concurrent requests cannot produce incorrect balances.
* The same idempotency key cannot create multiple payments.
* Temporary Temporal activity failures are retried automatically.
* Temporal recovery does not create duplicate financial transactions.
* Financial transactions are atomic.
* Payments do not remain indefinitely stuck in `PENDING`.
* Ambiguous API failures must not result in duplicate payments when clients retry using the same idempotency key.
