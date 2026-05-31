import pandas as pd

df = pd.DataFrame({
    "Runtime": [
        "FluxRuntime",
        "MutexQueue",
        "ChannelQueue",
    ],
    "Throughput_Mreqsec": [
        10.7,
        1.4,
        2.2,
    ],
    "P99_ns": [
        115,
        8000,
        4200,
    ],
})

latex = df.to_latex(
    index=False,
    float_format="%.2f",
)

with open(
    "papers/latex/tables/generated/comparison_table.tex",
    "w",
) as f:
    f.write(latex)

print("comparison_table complete")
