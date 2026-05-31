from graphviz import Digraph

g = Digraph(
    "FluxRuntime",
    format="png",
)

g.attr(rankdir="LR")

g.node("Ingress")
g.node("Dispatcher")
g.node("ShardQueues")
g.node("Workers")
g.node("Aggregation")
g.node("Redis")
g.node("Telemetry")

g.edge("Ingress", "Dispatcher")
g.edge("Dispatcher", "ShardQueues")
g.edge("ShardQueues", "Workers")
g.edge("Workers", "Aggregation")
g.edge("Aggregation", "Redis")
g.edge("Workers", "Telemetry")

g.render(
    "benchmarks/exported/figures/png/runtime_topology",
    cleanup=True,
)

print("runtime_topology complete")
