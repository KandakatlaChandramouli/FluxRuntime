import pandas as pd
import matplotlib.pyplot as plt
import numpy as np

latencies = pd.read_csv(
    "internal/workerpool/benchmarks/results/latencies.csv",
    header=None,
    names=["latency_ns"],
)

latencies["latency_us"] = latencies["latency_ns"] / 1000

p50 = np.percentile(latencies["latency_us"], 50)
p95 = np.percentile(latencies["latency_us"], 95)
p99 = np.percentile(latencies["latency_us"], 99)
p999 = np.percentile(latencies["latency_us"], 99.9)

print("\n=== Latency Statistics (microseconds) ===")
print(f"P50   : {p50:.2f} us")
print(f"P95   : {p95:.2f} us")
print(f"P99   : {p99:.2f} us")
print(f"P99.9 : {p999:.2f} us")

plt.figure(figsize=(10, 6))

plt.hist(
    latencies["latency_us"],
    bins=100,
)

plt.xlabel("Latency (microseconds)")
plt.ylabel("Frequency")
plt.title("Latency Distribution")

plt.savefig("benchmarks/figures/latency_histogram.png")

sorted_lat = np.sort(latencies["latency_us"])
percentiles = np.linspace(0, 100, len(sorted_lat))

plt.figure(figsize=(10, 6))

plt.plot(percentiles, sorted_lat)

plt.xlabel("Percentile")
plt.ylabel("Latency (microseconds)")
plt.title("Latency Percentile Curve")

plt.savefig("benchmarks/figures/latency_percentiles.png")

print("\nGraphs generated:")
print("- benchmarks/figures/latency_histogram.png")
print("- benchmarks/figures/latency_percentiles.png")
