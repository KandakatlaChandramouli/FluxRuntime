import matplotlib.pyplot as plt

x = [50, 95, 99]
y = [1.0, 2.5, 4.2]

plt.figure(figsize=(8,5))
plt.plot(x, y)
plt.xlabel("Percentile")
plt.ylabel("Latency")
plt.title("Latency Distribution")
plt.savefig("benchmarks/exported/plots/latency_distribution.png")
