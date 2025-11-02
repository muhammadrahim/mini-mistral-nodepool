APP_IMG ?= mini/job-runner:0.4.0
CLUSTER ?= mini-mistral

build:
	DOCKER_BUILDKIT=1 docker build --platform linux/arm64 -t $(APP_IMG) .

import:
	k3d image import -c $(CLUSTER) $(APP_IMG)

roll:
	kubectl -n app set image deploy/job-runner job-runner=$(APP_IMG)

apply-app:
	kubectl apply -f manifests/app/

apply-crd:
	kubectl apply -f manifests/crd/

apply-monitoring:
	kubectl apply -f manifests/monitoring/

forward:
	kubectl -n app port-forward svc/job-runner 8080:8080

logs-aibatch:
	@test -n "$(NAME)" || (echo "Usage: make logs-aibatch NAME=<aibatch-name>"; exit 1)
	@POD=$$(kubectl -n app get pod -l aibatch=$(NAME) -o jsonpath='{.items[0].metadata.name}'); \
	 NODE=$$(kubectl -n app get pod $$POD -o jsonpath='{.spec.nodeName}'); \
	 CID=$$(kubectl -n app get pod $$POD -o jsonpath='{.status.containerStatuses[0].containerID}' | sed 's!.*/!!'); \
	 echo "Node: $$NODE  Pod: $$POD  Container: $$CID"; \
	 docker exec -it $$NODE crictl logs $$CID