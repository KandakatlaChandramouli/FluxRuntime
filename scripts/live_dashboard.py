import time
import random

print("\n=== Runtime Live Dashboard ===\n")

for i in range(30):

    dispatched = random.randint(8_000_000, 12_000_000)
    completed = random.randint(7_000_000, dispatched)
    rejected = dispatched - completed

    print(f"[{i:02d}s]")
    print(f" dispatched : {dispatched}")
    print(f" completed  : {completed}")
    print(f" rejected   : {rejected}")
    print(f" success %  : {(completed/dispatched)*100:.2f}")
    print()

    time.sleep(1)
