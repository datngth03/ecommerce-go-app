import http from "k6/http";
import { check, sleep } from "k6";
import { Rate } from "k6/metrics";

const errorRate = new Rate("errors");
const BASE_URL = __ENV.BASE_URL || "http://localhost:8000";

export const options = {
   stages: [
      { duration: "2m", target: 100 }, // Ramp-up to 100 users
      { duration: "5m", target: 100 }, // Stay at 100 users
      { duration: "2m", target: 200 }, // Spike to 200 users
      { duration: "5m", target: 200 }, // Stay at 200 users
      { duration: "2m", target: 0 }, // Ramp-down to 0 users
   ],
   thresholds: {
      http_req_duration: ["p(95)<500", "p(99)<1000"],
      http_req_failed: ["rate<0.05"],
      errors: ["rate<0.1"],
   },
};

export default function () {
   // Test API Gateway health
   let healthRes = http.get(`${BASE_URL}/health`);
   check(healthRes, {
      "health check status is 200": (r) => r.status === 200,
   }) || errorRate.add(1);

   sleep(1);

   // Test product listing
   let productsRes = http.get(`${BASE_URL}/api/v1/products`, {
      headers: { "Content-Type": "application/json" },
   });
   check(productsRes, {
      "products status is 200": (r) => r.status === 200,
      "products response time < 500ms": (r) => r.timings.duration < 500,
   }) || errorRate.add(1);

   sleep(1);

   // Test inventory check
   const productId = "test-product-1";
   let inventoryRes = http.get(`${BASE_URL}/api/v1/inventory/stock/${productId}`);
   check(inventoryRes, {
      "inventory status is 200 or 404": (r) => r.status === 200 || r.status === 404,
   }) || errorRate.add(1);

   sleep(2);
}

export function handleSummary(data) {
   return {
      "load-test-results.json": JSON.stringify(data),
      stdout: textSummary(data, { indent: " ", enableColors: true }),
   };
}
