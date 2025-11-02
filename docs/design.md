# Mini Mistral Nodepool — Overview

```mermaid
flowchart LR
    A[AIBatch CR (ai.mini/v1)] -->|watch via informer| B[Controller]
    B -->|admission policy\n(tenant quota, priority)| C{Admit?}
    C -- Yes --> D[Job (batch/v1)]
    C -- No  --> E[Rejected\n(metric only)]
    D --> F[Pod (gpu=true)]
    F --> G[Workload (ai-sim)]
    G --> H[/Prometheus Metrics/]
    H --> I[Grafana]
```