KUBERNETES_VERSION=1.9
DIND_CLUSTER_SCRIPT=dind-cluster-v$(KUBERNETES_VERSION).sh
TEST_REGISTRY_PORT=50001
TEST_IMAGE_NAME=localhost:$(TEST_REGISTRY_PORT)/webhook-test

apply-webhook:
	@kubectl delete deployment k8s-admission-webhook || true
	cd test && \
	  ./create-signed-cert.sh --namespace default --service k8s-admission-webhook.default.svc && \
	  WEBHOOK_IMAGE=$(TEST_IMAGE_NAME) \
	  WEBHOOK_TLS_CERT=$$(cat server-cert.pem | sed 's/^/     /') \
	  WEBHOOK_TLS_PRIVATE_KEY_B64=$$(cat server-key.pem | base64 | tr -d '\n') \
	  WEBHOOK_CA_BUNDLE=$$(kubectl get configmap -n kube-system extension-apiserver-authentication -o=jsonpath='{.data.client-ca-file}' | base64 | tr -d '\n') \
	  envsubst < webhook.template.yaml > webhook.yaml && \
 	  rm csr.conf server-*.pem && \
	  kubectl apply -f webhook.yaml

dev-start: setup-test-cluster
dev-stop: cleanup-test-cluster cleanup-test-registry

setup-test-cluster:
	wget https://cdn.rawgit.com/kubernetes-sigs/kubeadm-dind-cluster/master/fixed/$(DIND_CLUSTER_SCRIPT) -O ./test/$(DIND_CLUSTER_SCRIPT)
	chmod +x ./test/$(DIND_CLUSTER_SCRIPT)
	APISERVER_admission_control=Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota \
	  NUM_NODES=1 \
	  ./test/$(DIND_CLUSTER_SCRIPT) up

cleanup-test-cluster:
	./test/$(DIND_CLUSTER_SCRIPT) down && ./test/$(DIND_CLUSTER_SCRIPT) clean
	@rm -rf ~/.kubeadm-dind-cluster || true

build-image-for-test:
	docker build -t $(TEST_IMAGE_NAME) .

setup-test-registry: cleanup-test-registry
	docker run -d -p $(TEST_REGISTRY_PORT):5000 --restart=always --name test-registry registry:2
	docker ps -a -q --filter=label=mirantis.kubeadm_dind_cluster | while read container_id; do \
	  docker exec $${container_id} /bin/bash -c "docker rm -fv registry-proxy || true"; \
	  docker exec $${container_id} /bin/bash -c "docker run --name registry-proxy -d -e LISTEN=':5000' -e TALK=\"\$$(/sbin/ip route|awk '/default/ { print \$$3 }'):$(TEST_REGISTRY_PORT)\" -p $(TEST_REGISTRY_PORT):5000 tecnativa/tcp-proxy"; \
	  done

cleanup-test-registry:
	@docker rm -fv test-registry || true

push-image-for-test: build-image-for-test setup-test-registry
	docker push $(TEST_IMAGE_NAME)
	docker rmi $(TEST_IMAGE_NAME)

deploy-webhook-for-test: push-image-for-test apply-webhook
