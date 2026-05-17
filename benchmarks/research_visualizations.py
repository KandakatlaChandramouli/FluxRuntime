import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
from pathlib import Path

Path("benchmarks/figures").mkdir(parents=True, exist_ok=True)


workers = [2, 4, 8, 16]

throughput = [15160.5, 15837, 15961.5, 15959]
dispatched = [19257.5, 19933.5, 20058, 20055.5]
rejected = [9064734, 9092457.5, 9144258.5, 9085687]
ns_op = [132, 130, 131, 131]

acceptance_rate = [
    (t / d) * 100 for t, d in zip(throughput, dispatched)
]

rejection_rate = [
    (r / (r + d)) * 100 for r, d in zip(rejected, dispatched)
]

scaling_efficiency = [
    throughput[i] / throughput[0]
    for i in range(len(throughput))
]

plt.figure(figsize=(10, 6))
plt.plot(workers, throughput, marker="o")
plt.xlabel("Worker Count")
plt.ylabel("Completed Requests")
plt.title("Throughput Scaling")
plt.grid(True)
plt.savefig("benchmarks/figures/research_throughput_scaling.png")

plt.figure(figsize=(10, 6))
plt.plot(workers, rejection_rate, marker="o")
plt.xlabel("Worker Count")
plt.ylabel("Rejection Rate (%)")
plt.title("Overload Rejection Behavior")
plt.grid(True)
plt.savefig("benchmarks/figures/research_rejection_rate.png")

plt.figure(figsize=(10, 6))
plt.plot(workers, ns_op, marker="o")
plt.xlabel("Worker Count")
plt.ylabel("ns/op")
plt.title("Dispatch Latency Scaling")
plt.grid(True)
plt.savefig("benchmarks/figures/research_latency_scaling.png")

plt.figure(figsize=(10, 6))
plt.plot(workers, scaling_efficiency, marker="o")
plt.xlabel("Worker Count")
plt.ylabel("Relative Throughput")
plt.title("Worker Scaling Efficiency")
plt.grid(True)
plt.savefig("benchmarks/figures/research_scaling_efficiency.png")

plt.figure(figsize=(10, 6))
plt.bar(
    ["Accepted", "Rejected"],
    [throughput[-1], rejected[-1]],
)
plt.ylabel("Requests")
plt.title("Queue Saturation Behavior")
plt.savefig("benchmarks/figures/research_queue_saturation.png")

print("\nGenerated research figures:")
print("- research_throughput_scaling.png")
print("- research_rejection_rate.png")
print("- research_latency_scaling.png")
print("- research_scaling_efficiency.png")
print("- research_queue_saturation.png")
