Grafana dashboards and guidance

Place exported dashboard JSON files here and reference them in deployment manifests.

Recommended dashboards:
- Service Overview (latency / error rate / throughput)
- Orders & Payments (orders per minute, payment failures)
- Inventory (stock levels, reservations, expirations)
- System (CPU, memory, network)

Provisioning:
- Use Grafana provisioning (configMaps in k8s) or import JSON directly via API.
