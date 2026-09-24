Login Load Test — Final
Target: 10 RPS
Test: 25 concurrent users, 30 seconds
Observed throughput: 50–179 RPS across repeated runs
Latency: roughly 140–495 ms average depending on run
Errors: 0
HTTP 200: 100% successful
Conclusion: Login comfortably meets the 10 RPS target under the current local setup.

We also tested wrong-password requests and they correctly returned 401, with similar performance.

Login performance → done.


Refresh Token Load Test — Summary
Concurrency: 10
Duration: 30 seconds
Requests: 23,239
Throughput: 774 RPS
Average latency: 12.87 ms
P50: 12.03 ms
P95: 21.80 ms
P99: 33.10 ms
Success rate: 100%
HTTP 200: 23,239
Target: ~25 RPS

Conclusion: Refresh-token flow comfortably handles the expected SamPay workload with very low latency and zero failures. No optimization needed at this stage.


Forgot Password Load Test — Summary

Concurrency: 10
Duration: 10 seconds
Requests: 2,542
Throughput: 254 RPS
Average latency: 39.32 ms
P50: 37.28 ms
P95: 76.66 ms
P99: 95.13 ms
Success rate: 100%
HTTP 200: 2,542
Target: ~25 RPS

Conclusion: Forgot-password flow comfortably handles the expected SamPay workload with low HTTP latency and zero failures. Load testing also exposed and helped fix a duplicate reset-token generation issue under high concurrency. No further optimization needed at this stage.


Reset Password Load Test — Summary

Concurrency: 10
Duration: 10 seconds
Requests: 1,000
Throughput: 135.95 RPS
Average latency: 73.24 ms
P50: 66.02 ms
P95: 117.59 ms
P99: 147.37 ms
Success rate: 100%
HTTP 200: 1,000
Target: ~25 RPS

Conclusion: Reset-password flow comfortably handles the expected SamPay workload with low latency and zero failures. The benchmark also exercised bcrypt password hashing under concurrent load. No further optimization needed at this stage.