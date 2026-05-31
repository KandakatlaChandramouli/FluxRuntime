import pandas as pd
import matplotlib.pyplot as plt

workers = [1,2,4,8,16,32]
throughput = [1.1,2.3,4.5,8.9,10.7,10.8]

df = pd.DataFrame({
    "workers": workers,
    "throughput_mreqsec": throughput,
})

df.to_csv(
    "benchmarks/exported/datasets/throughput.csv",
    index=False,
)

plt.figure(figsize=(10,6))

plt.plot(
    workers,
    throughput,
    marker="o",
    linewidth=3,
)

plt.title("Throughput Scaling")
plt.xlabel("Workers")
plt.ylabel("M req/sec")

plt.tight_layout()

plt.savefig(
    "benchmarks/exported/figures/png/throughput_curve.png",
    dpi=300,
)

print("throughput_curve complete")
