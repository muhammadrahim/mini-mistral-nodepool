package controller

import (
	"context"
	"fmt"
	"log"

	"github.com/muhammadrahim/mini-mistral-nodepool/pkg/httpx"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
)

func str(u *unstructured.Unstructured, path ...string) string {
	v, _, _ := unstructured.NestedString(u.Object, path...)
	return v
}
func i64(u *unstructured.Unstructured, path ...string) int64 {
	v, _, _ := unstructured.NestedInt64(u.Object, path...)
	return v
}
func b32(i int32) *int32 { return &i }
func bptr(b bool) *bool  { return &b }

func Reconcile(ctx context.Context, cs *kubernetes.Clientset, u *unstructured.Unstructured) {
	ns, name := u.GetNamespace(), u.GetName()
	tenant := str(u, "spec", "tenant")
	workload := str(u, "spec", "workload")
	prio := str(u, "spec", "priorityClass")
	tokens := i64(u, "spec", "tokensTargetPerSec")
	batchSz := i64(u, "spec", "batchSize")

	// Tenant concurrency quota: at most 3 Active
	jl, _ := cs.BatchV1().Jobs(ns).List(ctx, meta.ListOptions{LabelSelector: fmt.Sprintf("tenant=%s", tenant)})
	active := 0
	for _, j := range jl.Items {
		if j.Status.Active > 0 { active++ }
	}
	if active >= 3 {
		httpx.AIBatchRejected.WithLabelValues(tenant, "quota").Inc()
		log.Printf("AIBatch %s/%s rejected: quota", ns, name)
		return
	}

	jobName := "aibatch-" + name
	if _, err := cs.BatchV1().Jobs(ns).Get(ctx, jobName, meta.GetOptions{}); err == nil {
		return // already exists
	}

	job := &batchv1.Job{
		ObjectMeta: meta.ObjectMeta{
			Name: jobName, Namespace: ns,
			Labels: map[string]string{"app":"mini-mistral","aibatch":name,"tenant":tenant,"workload":workload},
			OwnerReferences: []meta.OwnerReference{{
				APIVersion: u.GetAPIVersion(), Kind: u.GetKind(), Name: u.GetName(), UID: u.GetUID(),
				Controller: bptr(true), BlockOwnerDeletion: bptr(true),
			}},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: b32(1),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{Labels: map[string]string{"app":"mini-mistral","aibatch":name,"tenant":tenant,"workload":workload}},
				Spec: corev1.PodSpec{
					NodeSelector:      map[string]string{"gpu":"true"},
					PriorityClassName: prio,
					RestartPolicy:     corev1.RestartPolicyNever,
					Containers: []corev1.Container{{
						Name:  "ai-sim",
						Image: "public.ecr.aws/docker/library/alpine:3.20",
						Command: []string{"sh","-c", fmt.Sprintf(
							"echo starting %s tenant=%s; sleep 2; echo tokens/s=%d batch=%d; sleep 10; echo done",
							workload, tenant, tokens, batchSz,
						)},
					}},
				},
			},
		},
	}
	if _, err := cs.BatchV1().Jobs(ns).Create(ctx, job, meta.CreateOptions{}); err != nil {
		log.Printf("create Job for AIBatch %s/%s: %v", ns, name, err)
		return
	}
	httpx.AIBatchAdmitted.WithLabelValues(tenant, workload).Inc()
	log.Printf("AIBatch %s/%s admitted -> Job %s", ns, name, job.Name)
}
