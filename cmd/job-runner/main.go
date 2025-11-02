package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/muhammadrahim/mini-mistral-nodepool/pkg/controller"
	"github.com/muhammadrahim/mini-mistral-nodepool/pkg/httpx"
	"github.com/muhammadrahim/mini-mistral-nodepool/pkg/kube"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	prometheus.MustRegister(httpx.JobsSubmitted, httpx.JobSubmitErrors, httpx.AIBatchAdmitted, httpx.AIBatchRejected)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cs := kube.MustClient()
	cfg, _ := kube.Config()
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil { log.Fatalf("dynamic client: %v", err) }

	// Start AIBatch informer (namespace "app")
	stopCh := controller.StartInformer(dyn, "app", func(obj interface{}) {
		u := obj.(*unstructured.Unstructured) // import if needed
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		controller.Reconcile(ctx, cs, u)
	})

	mux := httpx.NewMux(cs)
	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{ Addr: ":8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second }
	log.Println("job-runner listening on :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		close(stopCh)
		log.Fatal(err)
	}
}
