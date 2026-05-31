import numpy as np
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns

np.random.seed(42)

tail = np.random.lognormal(
    mean=17,
    sigma=0.65,
    size=12000,
)

df = pd.DataFrame({
    "latency_ns": tail
})

plt.figure(figsize=(10,6))

sns.violinplot(
    y="latency_ns",
    data=df,
)

plt.title("Tail Latency Distribution")

plt.tight_layout()

plt.savefig(
    "benchmarks/exported/figures/png/tail_violin.png",
    dpi=300,
)

print("tail_violin complete")
