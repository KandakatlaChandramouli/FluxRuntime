import matplotlib.pyplot as plt

load = [10,20,30,40,50,60,70,80]
reject = [0,1,2,4,8,16,32,48]

plt.figure(figsize=(10,6))

plt.plot(
    load,
    reject,
    marker="o",
    linewidth=3,
)

plt.title("Overload Rejection Dynamics")
plt.xlabel("Ingress Load")
plt.ylabel("Rejected %")

plt.tight_layout()

plt.savefig(
    "benchmarks/exported/figures/png/rejection_curve.png",
    dpi=300,
)

print("rejection_curve complete")
