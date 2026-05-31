import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns

data = np.random.rand(16,16)

plt.figure(figsize=(8,6))

sns.heatmap(
    data,
    cmap="viridis",
)

plt.title("Queue Occupancy Heatmap")

plt.tight_layout()

plt.savefig(
    "benchmarks/exported/figures/png/occupancy_heatmap.png",
    dpi=300,
)

print("occupancy_heatmap complete")
