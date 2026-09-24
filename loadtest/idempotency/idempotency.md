# Idempotency Load Test

## Objective

Verify that multiple concurrent requests using the same `Idempotency-Key` do not create duplicate payments.

## Test

* Send multiple concurrent payment requests.
* All requests use the same `Idempotency-Key`.
* The payment amount and request payload are identical.
* Verify that only one payment is actually created.

## Expected Behavior

* One request successfully creates and processes the payment.
* Concurrent duplicate requests receive `409 REQUEST_IN_PROGRESS` while the original request is still processing.
* Once the original request completes, subsequent requests using the same key return the stored response.
* No duplicate payment should be created.

## Result

The test passed.

For 10 concurrent requests using the same idempotency key:

* **1 payment** was created.
* Concurrent requests received `409 REQUEST_IN_PROGRESS`.
* A subsequent retry received the same successful payment response.
* No duplicate payment was created.

## Conclusion

The idempotency mechanism correctly prevents duplicate financial operations when the same request is submitted multiple times concurrently or retried after completion.
