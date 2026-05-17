import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
import re

CSV_PATH = "benchmarks/results/pressure_matrix.csv"
OUTPUT_DIR = "benchmarks/figures"

with open(CSV_PATH, "r") as f:
    lines = f.readlines()

sections = {}
current_metric = None
rows = []

for line in lines:
    line = line.strip()

    if not line:
        continue

    if line.startswith(",") and "CI" in line:
        current_metric = line.split(",")[1]
        sections[current_metric] = []
        continue

    if line.startswith("PressureMatrix"):
        parts = line.split(",")
        name = parts[0]
        value = parts[1]

        match = re.search(r"workers_(\d+)", name)
        if not match:
            continue

        workers = int(match.group(1))

        try:
            value = float(value)
        except:
            continue

        sections[current_metric].append((workers, value))


def plot_metric(metric_name, ylabel, filename, scale=1.0):
    data = sorted(sections.get(metric_name, []))

    if not data:
        print(f"No data for {metric_name}")
        return

    workers = [x[0] for x in data]
    values = [x[1] * scale for x in data]

    plt.figure(figsize=(10, 6))

    plt.plot(
        workers,
        values,
        marker="o",
        linewidth=3,
        markersize=10,
    )

    plt.xticks(workers, fontsize=12)
    plt.yticks(fontsize=12)

    plt.xlabel(
        "Worker Count",
        fontsize=16,
        fontweight="bold",
    )

    plt.ylabel(
        ylabel,
        fontsize=16,
        fontweight="bold",
    )

    plt.title(
        f"{metric_name} Scaling Behavior",
        fontsize=20,
        fontweight="bold",
        pad=20,
    )

    plt.grid(True, linestyle="--", alpha=0.6)

    # Annotate points
    for x, y in zip(workers, values):
        plt.annotate(
            f"{y:.0f}",
            (x, y),
            textcoords="offset points",
            xytext=(0, 10),
            ha="center",
            fontsize=11,
        )

    plt.tight_layout()

    plt.savefig(
        f"{OUTPUT_DIR}/{filename}",
        dpi=400,
        bbox_inches="tight",
    )

    plt.close()
