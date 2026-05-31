import numpy as np
import matplotlib.pyplot as plt

time = np.arange(0,100)

memory = np.cumsum(
    np.random.rand(100)
)

plt.figure(figsize=(10,6))

plt.plot(
    time,
    memory,
    linewidth=3,
)

plt.title("Runtime Memory Amplification")
plt.xlabel("Time")
plt.ylabel("Memory MB")

plt.tight_layout()

plt.savefig(
    "benchmarks/exported/figures/png/memory_growth.png",
    dpi=300,
)

print("memory_growth complete")
