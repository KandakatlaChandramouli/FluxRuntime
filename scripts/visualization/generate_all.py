import subprocess

scripts = [
    "scripts/visualization/latency/latency_distribution.py",
    "scripts/visualization/scaling/core_scaling.py",
    "scripts/visualization/publication/queue_comparison.py",
]

for s in scripts:

    subprocess.run(
        ["python3", s],
        check=True,
    )

print("ALL FIGURES GENERATED")
