package controller

import (
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/muhammadrahim/mini-mistral-nodepool/pkg/httpx"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

var AIBatchGVR = schema.GroupVersionResource{Group: "ai.mini", Version: "v1", Resource: "aibatches"}

func RegisterMetrics() {
	prometheus.MustRegister(httpx.AIBatchAdmitted, httpx.AIBatchRejected, httpx.JobsSubmitted, httpx.JobSubmitErrors)
}

func StartInformer(dyn dynamic.Interface, ns string, onSync func(obj interface{})) (stopCh chan struct{}) {
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dyn, 0, ns, nil)
	inf := factory.ForResource(AIBatchGVR).Informer()
	inf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onSync,
		UpdateFunc: func(_, newObj interface{}) { onSync(newObj) },
	})
	stopCh = make(chan struct{})
	factory.Start(stopCh)
	if ok := cache.WaitForCacheSync(stopCh, inf.HasSynced); !ok {
		log.Fatal("AIBatch informer cache failed to sync")
	}
	return stopCh
}

func RequeueAfter(d time.Duration) { time.Sleep(d) }
