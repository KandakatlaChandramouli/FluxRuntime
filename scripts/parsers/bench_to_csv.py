import re
import pandas as pd

pattern = re.compile(
    r"^(Benchmark\S+)\s+\d+\s+([\d.]+)\sns/op\s+(\d+)\sB/op\s+(\d+)\sallocs/op"
)

rows = []

with open("benchmarks/raw/full_benchmark.txt") as f:

    for line in f:

        m = pattern.search(line)

        if m:

            rows.append({
                "benchmark": m.group(1),
                "ns_op": float(m.group(2)),
                "b_op": int(m.group(3)),
                "allocs_op": int(m.group(4)),
            })

df = pd.DataFrame(rows)

df.to_csv(
    "benchmarks/processed/benchmarks.csv",
    index=False,
)

print(df.head())
