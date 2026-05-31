import pandas as pd
import matplotlib.pyplot as plt
from pathlib import Path

Path("benchmarks/figures").mkdir(parents=True, exist_ok=True)

df = pd.read_csv("internal/workerpool/benchmarks/results/latencies.csv")

plt.figure(figsize=(12,6))
for col in df.columns[:4]:
    plt.plot(df.index, df[col], label=col)

plt.xlabel("Samples")
plt.ylabel("Latency (ns)")
plt.title("FluxRuntime Latency Evolution")
plt.legend()
plt.grid(True)

plt.savefig("benchmarks/figures/runtime_latency.png")
