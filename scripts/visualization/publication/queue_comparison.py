import matplotlib.pyplot as plt

queues = [
    "mutex",
    "channel",
    "unbounded",
]

latencies = [
    8.6,
    10.8,
    392,
]

plt.figure(figsize=(8,5))

plt.bar(
    queues,
    latencies,
)

plt.ylabel("ns/op")
plt.title("Queue Baseline Comparison")

plt.tight_layout()

plt.savefig(
    "benchmarks/exported/figures/png/queue_comparison.png",
    dpi=300,
)

print("queue_comparison complete")
