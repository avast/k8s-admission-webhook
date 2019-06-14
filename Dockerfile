FROM golang:1.10-stretch AS builder

RUN curl https://glide.sh/get | sh

WORKDIR /go/src/github.com/avast/k8s-admission-webhook

COPY glide.* /go/src/github.com/avast/k8s-admission-webhook/
RUN glide install -v

COPY . . 

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s"

# ---
FROM scratch

LABEL org.label-schema.name="k8s-admission-webhook"
LABEL org.label-schema.description="General-purpose Kubernetes admission webhook"
LABEL org.label-schema.vcs-url="https://github.com/avast/k8s-admission-webhook"
LABEL org.label-schema.vendor="Avast Software"

WORKDIR /app

COPY --from=builder /go/src/github.com/avast/k8s-admission-webhook/k8s-admission-webhook .

ENTRYPOINT ["./k8s-admission-webhook", "webhook"]