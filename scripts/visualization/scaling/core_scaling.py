import matplotlib.pyplot as plt

workers = [1,2,4,8,16]
throughput = [1.2,2.4,4.8,9.7,10.9]

plt.figure(figsize=(8,5))

plt.plot(
    workers,
    throughput,
    marker="o",
)

plt.xlabel("Workers")
plt.ylabel("Throughput (M req/sec)")
plt.title("FluxRuntime Throughput Scaling")

plt.tight_layout()

plt.savefig(
    "benchmarks/exported/figures/png/core_scaling.png",
    dpi=300,
)

print("core_scaling complete")
