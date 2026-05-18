import pandas as pd
import matplotlib.pyplot as plt
from pathlib import Path

Path("benchmarks/figures").mkdir(parents=True, exist_ok=True)

# Scaling benchmark
workers = [2, 4, 8, 16]
throughput = [5.5, 6.4, 10.3, 11.4]
latency = [211.3, 178.0, 116.3, 106.2]

plt.figure(figsize=(8,5))
plt.plot(workers, throughput, marker='o')
plt.xlabel("Worker Count")
plt.ylabel("Throughput (Million req/sec)")
plt.title("Worker Scaling Throughput")
plt.grid(True)
plt.savefig("benchmarks/figures/scaling_throughput.png")
plt.close()

plt.figure(figsize=(8,5))
plt.plot(workers, latency, marker='o')
plt.xlabel("Worker Count")
plt.ylabel("Latency (ns/op)")
plt.title("Worker Scaling Latency")
plt.grid(True)
plt.savefig("benchmarks/figures/scaling_latency.png")
plt.close()

# Latency telemetry
csv_path = "internal/workerpool/benchmarks/results/latencies.csv"

try:
    df = pd.read_csv(csv_path)

    plt.figure(figsize=(10,5))

    for col in df.columns[:4]:
        plt.plot(df.index, df[col], label=col)

    plt.xlabel("Sample")
    plt.ylabel("Latency")
    plt.title("Runtime Latency Evolution")
    plt.legend()
    plt.grid(True)

    plt.savefig("benchmarks/figures/runtime_latency_evolution.png")
    plt.close()

except Exception as e:
    print("Could not generate latency evolution chart:", e)

print("Charts generated successfully.")
