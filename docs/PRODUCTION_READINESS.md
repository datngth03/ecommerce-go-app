# Production Readiness Checklist

This document summarizes steps to make the e-commerce platform production-grade.

## Observability

- Prometheus: scrape endpoints on `/metrics` for all services. See `infrastructure/monitoring/prometheus2/prometheus.yml`.
- Grafana: import dashboards (place under `infrastructure/monitoring/grafana/`).
- Logs: centralize logs (ELK, Loki, or Cloud provider logging). Ensure logs are structured JSON and include correlation IDs.

## Distributed Tracing

- Instrument services with OpenTelemetry and export to Jaeger.
- Sample collector: deploy `otel-collector` as a daemonset and configure OTLP exporter.

## CI/CD

- GitHub Actions workflow added at `.github/workflows/ci-cd.yml` for build/test and image push.
- Use image signing and vulnerability scanning (Trivy) in pipeline.

## Security Hardening

- Secrets: store in Vault / Kubernetes Secrets and avoid plaintext in compose files.
- TLS: enforce TLS for all service-to-service traffic (mTLS for internal comms).
- Rate limiting and authentication at API Gateway.
- Regular dependency scanning and patching.

## Performance Optimization

- Database: proper indexes (already present), connection pool tuning, read-replicas for scale.
- Caching: Redis caches for hot endpoints (already used in inventory).
- Asynchronous: move long work to background jobs and message queues (RabbitMQ).

## Rollout Strategy

- Blue/Green or Canary deployments in Kubernetes.
- Health checks and readiness probes implemented in services.

## Next Steps

1. Add OpenTelemetry instrumentation to each service (example PR template provided on request).
2. Deploy Prometheus + Grafana + Jaeger in k8s manifests (I can scaffold these).
3. Add CI steps: container scan, vulnerability alerts, and deployment approvals.
