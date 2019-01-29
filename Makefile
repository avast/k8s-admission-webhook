KUBERNETES_VERSION=1.9
DIND_CLUSTER_SCRIPT=dind-cluster-v$(KUBERNETES_VERSION).sh
TEST_IMAGE_NAME=webhook-test:latest
KUBECTL=~/.kubeadm-dind-cluster/kubectl
ADMISSION_PLUGIN_LIST=Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota

dev-start: setup-test-cluster
dev-stop: cleanup-test-cluster

dev-e2e-test: deploy-webhook-for-test
	KUBECTL=$(KUBECTL) go test -tags=e2e -p 1

ci-e2e-test: setup-test-cluster deploy-webhook-for-test
	sleep 20 && \
	  $(KUBECTL) describe deployment k8s-admission-webhook --namespace=default && \
	  KUBECTL=$(KUBECTL) go test -tags=e2e -p 1

apply-webhook:
	@$(KUBECTL) delete deployment k8s-admission-webhook || true
	cd test && \
	  export PATH=~/.kubeadm-dind-cluster:$$PATH && \
	  	./create-signed-cert.sh --namespace default --service k8s-admission-webhook.default.svc && \
	  WEBHOOK_IMAGE=$(TEST_IMAGE_NAME) \
	  WEBHOOK_TLS_CERT=$$(cat server-cert.pem | sed 's/^/     /') \
	  WEBHOOK_TLS_PRIVATE_KEY_B64=$$(cat server-key.pem | base64 | tr -d '\n') \
	  WEBHOOK_CA_BUNDLE=$$($(KUBECTL) get configmap -n kube-system extension-apiserver-authentication -o=jsonpath='{.data.client-ca-file}' | base64 | tr -d '\n') \
	  envsubst < webhook.template.yaml > webhook.yaml && \
 	  rm csr.conf server-*.pem && \
	  $(KUBECTL) apply -f webhook.yaml

setup-test-cluster:
	if [ "$(KUBERNETES_VERSION)" = "1.9" ]; then \
		wget https://raw.githubusercontent.com/kubernetes-sigs/kubeadm-dind-cluster/3cde76608aed0d64895077a0cf2f2e3b9e7323da/fixed/dind-cluster-v1.9.sh -O ./test/$(DIND_CLUSTER_SCRIPT); \
	else \
		wget https://github.com/kubernetes-sigs/kubeadm-dind-cluster/releases/download/v0.1.0/$(DIND_CLUSTER_SCRIPT) -O ./test/$(DIND_CLUSTER_SCRIPT); \
	fi
	chmod +x ./test/$(DIND_CLUSTER_SCRIPT)
	export API_SERVER_ADMISSION_ARG=$$(if [ "$(KUBERNETES_VERSION)" = "1.11" ] || [ "$(KUBERNETES_VERSION)" = "1.12" ] || [ "$(KUBERNETES_VERSION)" = "1.13" ]; then echo enable_admission_plugins; else echo admission_control; fi) && \
	  export APISERVER_$${API_SERVER_ADMISSION_ARG}=$(ADMISSION_PLUGIN_LIST) && \
	  NUM_NODES=1 \
	    ./test/$(DIND_CLUSTER_SCRIPT) up

cleanup-test-cluster:
	./test/$(DIND_CLUSTER_SCRIPT) down && ./test/$(DIND_CLUSTER_SCRIPT) clean
	@rm -rf ~/.kubeadm-dind-cluster || true

build-image-for-test:
	docker build -t $(TEST_IMAGE_NAME) .

copy-image-for-test: build-image-for-test
	./test/$(DIND_CLUSTER_SCRIPT) copy-image $(TEST_IMAGE_NAME)
	docker rmi $(TEST_IMAGE_NAME)

deploy-webhook-for-test: copy-image-for-test apply-webhook
