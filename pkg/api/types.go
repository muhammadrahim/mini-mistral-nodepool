package api

type AIBatchSpec struct {
	Tenant             string `json:"tenant"`
	ModelHint          string `json:"modelHint"`
	Workload           string `json:"workload"` // inference|training
	GPUCount           int    `json:"gpuCount"`
	BatchSize          int    `json:"batchSize"`
	TokensTargetPerSec int    `json:"tokensTargetPerSec"`
	PriorityClass      string `json:"priorityClass"`
}