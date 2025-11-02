package httpx

import "github.com/prometheus/client_golang/prometheus"

var (
	JobsSubmitted = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "mini", Subsystem: "jobrunner", Name: "jobs_submitted_total",
		Help: "Total number of jobs submitted.",
	})
	JobSubmitErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "mini", Subsystem: "jobrunner", Name: "job_submit_errors_total",
		Help: "Total number of errors when submitting jobs.",
	})
	AIBatchAdmitted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "mini", Subsystem: "controller", Name: "aibatch_admitted_total",
			Help: "AIBatches admitted (Job created).",
		}, []string{"tenant","workload"},
	)
	AIBatchRejected = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "mini", Subsystem: "controller", Name: "aibatch_rejected_total",
			Help: "AIBatches rejected (e.g., quota).",
		}, []string{"tenant","reason"},
	)
)
