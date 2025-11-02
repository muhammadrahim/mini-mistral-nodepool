```mermaid
classDiagram
    class AIBatch {
      +tenant: string
      +modelHint: string
      +workload: string
      +gpuCount: int
      +batchSize: int
      +tokensTargetPerSec: int
      +priorityClass: string
    }

    class Controller {
      +StartInformer(ns)
      +Reconcile(AIBatch)
      -listJobsByTenant()
      -makeJobFromAIBatch()
    }

    class JobRunnerService {
      +/submit
      +/status
      +/metrics
    }

    class Metrics {
      +jobs_submitted_total
      +job_submit_errors_total
      +aibatch_admitted_total{tenant,workload}
      +aibatch_rejected_total{tenant,reason}
    }

    AIBatch <.. Controller : watches
    Controller --> Job : creates
    Job --> Pod
    JobRunnerService ..> Metrics : exposes
    Controller ..> Metrics : exposes
```