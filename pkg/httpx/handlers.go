package httpx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

type SubmitReq struct {
	DurationSeconds int    `json:"duration_seconds"`
	NamePrefix      string `json:"name_prefix"`
	Namespace       string `json:"namespace"`
}

type SubmitResp struct {
	JobName   string `json:"job_name"`
	Namespace string `json:"namespace"`
}

func int32p(i int32) *int32 { return &i }

func NewMux(cs *kubernetes.Clientset) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/submit", submitHandler(cs))
	mux.HandleFunc("/status", statusHandler(cs))
	return mux
}

func submitHandler(cs *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SubmitReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if req.DurationSeconds <= 0 { req.DurationSeconds = 10 }
		if req.Namespace == "" { req.Namespace = "default" }

		name := fmt.Sprintf("%s-%04d", pickPrefix(req.NamePrefix), rand4())
		job := &batchv1.Job{
			ObjectMeta: meta.ObjectMeta{
				Name:      name, Namespace: req.Namespace,
				Labels:    map[string]string{"app":"mini-mistral"},
			},
			Spec: batchv1.JobSpec{
				BackoffLimit: int32p(1),
				Template: corev1.PodTemplateSpec{
					ObjectMeta: meta.ObjectMeta{Labels: map[string]string{"app":"mini-mistral"}},
					Spec: corev1.PodSpec{
						NodeSelector:  map[string]string{"gpu":"true"},
						RestartPolicy: corev1.RestartPolicyNever,
						Containers: []corev1.Container{{
							Name: "simulate-ai",
							Image: "public.ecr.aws/docker/library/alpine:3.20",
							Command: []string{"sh","-c", fmt.Sprintf("echo training...; sleep %d; echo done", req.DurationSeconds)},
							Ports: []corev1.ContainerPort{{ContainerPort: 8080, Protocol: corev1.ProtocolTCP}},
							ReadinessProbe: &corev1.Probe{ProbeHandler: corev1.ProbeHandler{
								Exec: &corev1.ExecAction{Command: []string{"sh","-c","echo ok"}},
							}, InitialDelaySeconds:1, TimeoutSeconds:1, PeriodSeconds:5, FailureThreshold:3},
							LivenessProbe: &corev1.Probe{ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{Path:"/healthz", Port:intstr.FromInt(8080)},
							}, InitialDelaySeconds:5, TimeoutSeconds:1, PeriodSeconds:10, FailureThreshold:3},
						}},
					},
				},
			},
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second); defer cancel()
		if _, err := cs.BatchV1().Jobs(req.Namespace).Create(ctx, job, meta.CreateOptions{}); err != nil {
			JobSubmitErrors.Inc()
			http.Error(w, fmt.Sprintf("create job: %v", err), http.StatusInternalServerError)
			return
		}
		JobsSubmitted.Inc()
		w.Header().Set("Content-Type","application/json")
		_ = json.NewEncoder(w).Encode(SubmitResp{JobName: name, Namespace: req.Namespace})
	}
}

func statusHandler(cs *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		ns := r.URL.Query().Get("ns")
		if ns == "" { ns = "default" }
		if name == "" {
			http.Error(w, "missing ?name=", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second); defer cancel()
		job, err := cs.BatchV1().Jobs(ns).Get(ctx, name, meta.GetOptions{})
		if err != nil {
			http.Error(w, fmt.Sprintf("get job: %v", err), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type","application/json")
		_ = json.NewEncoder(w).Encode(job.Status)
	}
}

func pickPrefix(p string) string {
	if p != "" { return p }
	return "ai-job"
}
func rand4() int { return int(time.Now().UnixNano()%10000) }
