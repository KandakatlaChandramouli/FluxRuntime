import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns

df = pd.read_csv(
    "benchmarks/processed/benchmarks.csv"
)

plt.figure(figsize=(10,6))

sns.barplot(
    data=df,
    x="benchmark",
    y="ns_op",
)

plt.xticks(rotation=70)

plt.title("Benchmark Latency Distribution")
plt.ylabel("ns/op")

plt.tight_layout()

plt.savefig(
    "benchmarks/exported/figures/png/latency_distribution.png",
    dpi=300,
)

print("latency_distribution complete")
