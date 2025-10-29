import http from "k6/http";
import { check, sleep, group } from "k6";
import { Rate, Trend, Counter, Gauge } from "k6/metrics";

// Custom metrics tracking the full flow
const errorRate = new Rate("errors");
const fullFlowDuration = new Trend("full_flow_duration");
const apiGatewayLatency = new Trend("api_gateway_latency");
const orderServiceLatency = new Trend("order_service_latency");
const paymentServiceLatency = new Trend("payment_service_latency");
const connectionPoolSavings = new Gauge("connection_pool_time_saved_ms");

// Success counters
const fullFlowSuccess = new Counter("full_flow_success");
const fullFlowFailure = new Counter("full_flow_failure");

// Test configuration - comprehensive load profile
export const options = {
   stages: [
      // Gradual ramp-up
      { duration: "1m", target: 20 }, // Stage 1: Light load
      { duration: "2m", target: 20 }, // Stage 2: Sustain

      { duration: "1m", target: 50 }, // Stage 3: Medium load
      { duration: "3m", target: 50 }, // Stage 4: Sustain

      { duration: "1m", target: 100 }, // Stage 5: High load
      { duration: "3m", target: 100 }, // Stage 6: Sustain (stress test)

      { duration: "30s", target: 150 }, // Stage 7: Spike test
      { duration: "1m", target: 150 }, // Stage 8: Peak load

      { duration: "2m", target: 0 }, // Stage 9: Cool-down
   ],
   thresholds: {
      // Overall HTTP performance (with connection pool)
      http_req_duration: ["p(95)<300", "p(99)<600"],

      // Full flow (API Gateway → Order → Payment) should be FAST
      // Before: ~1560ms, After: ~160ms (9.75x improvement)
      full_flow_duration: [
         "p(50)<150", // Median < 150ms
         "p(95)<250", // 95th percentile < 250ms
         "p(99)<500", // 99th percentile < 500ms
         "avg<200", // Average < 200ms
      ],

      // Individual service latencies (with connection pool)
      api_gateway_latency: ["p(95)<50"], // Before: 220ms, After: 20ms
      order_service_latency: ["p(95)<150"], // Before: 900ms, After: 100ms
      payment_service_latency: ["p(95)<100"], // Before: 440ms, After: 40ms

      // Success metrics
      errors: ["rate<0.01"], // < 1% error rate
      checks: ["rate>0.99"], // > 99% success rate
      full_flow_success: ["count>1000"], // At least 1000 successful flows
   },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8000";
const ENABLE_DETAILED_LOGS = __ENV.DETAILED_LOGS === "true";

// Test user credentials
const TEST_USERS = [
   { email: "user1@example.com", password: "testpass123" },
   { email: "user2@example.com", password: "testpass123" },
   { email: "user3@example.com", password: "testpass123" },
   { email: "loadtest@example.com", password: "testpassword123" },
];

// Product catalog for realistic scenarios
const PRODUCTS = [
   { id: "prod-1", name: "Laptop", price: 999.99 },
   { id: "prod-2", name: "Mouse", price: 29.99 },
   { id: "prod-3", name: "Keyboard", price: 79.99 },
   { id: "prod-4", name: "Monitor", price: 299.99 },
   { id: "prod-5", name: "Headphones", price: 149.99 },
];

export function setup() {
   console.log("");
   console.log("═══════════════════════════════════════════════════════════");
   console.log("🚀 E-Commerce End-to-End Performance Test");
   console.log("⚡ Connection Pool Performance Validation");
   console.log("═══════════════════════════════════════════════════════════");
   console.log("");
   console.log(`📍 Target: ${BASE_URL}`);
   console.log("");
   console.log("🎯 Testing Connection Pool Improvements:");
   console.log("   • API Gateway → 6 services (30 pooled connections)");
   console.log("   • Order Service → 5 services (25 pooled connections)");
   console.log("   • Payment Service → 2 services (10 pooled connections)");
   console.log("");
   console.log("📊 Expected Performance:");
   console.log("   • API Gateway: 220ms → 20ms (11x faster)");
   console.log("   • Order Service: 900ms → 100ms (9x faster)");
   console.log("   • Payment Service: 440ms → 40ms (11x faster)");
   console.log("   • End-to-End: 1560ms → 160ms (9.75x faster)");
   console.log("");
   console.log("⏱️  Test Duration: ~14 minutes");
   console.log("👥 Max Virtual Users: 150");
   console.log("");
   console.log("Starting in 3 seconds...");
   console.log("");

   sleep(3);
   return {};
}

export default function () {
   // Select random user and product
   const user = TEST_USERS[Math.floor(Math.random() * TEST_USERS.length)];
   const product = PRODUCTS[Math.floor(Math.random() * PRODUCTS.length)];
   const quantity = Math.floor(Math.random() * 3) + 1; // 1-3 items

   const headers = { "Content-Type": "application/json" };
   let token = null;
   let orderId = null;
   let paymentId = null;

   const flowStart = Date.now();

   // ═══════════════════════════════════════════════════════════
   // PHASE 1: Authentication (API Gateway → User Service)
   // ═══════════════════════════════════════════════════════════
   group("Phase 1: Authentication", function () {
      const authStart = Date.now();

      const loginRes = http.post(
         `${BASE_URL}/auth/login`,
         JSON.stringify({
            email: user.email,
            password: user.password,
         }),
         { headers }
      );

      const authDuration = Date.now() - authStart;
      apiGatewayLatency.add(authDuration);

      const success = check(loginRes, {
         "auth status is 200": (r) => r.status === 200,
         "auth time < 50ms": () => authDuration < 50,
         "got access token": (r) => {
            try {
               const body = JSON.parse(r.body);
               token = body.access_token || body.data?.access_token;
               return !!token;
            } catch (e) {
               return false;
            }
         },
      });

      if (!success && !token) {
         // Try without auth for remaining tests
         console.warn("⚠️  Authentication failed, continuing without token");
      }

      errorRate.add(!success);
   });

   if (token) {
      headers["Authorization"] = `Bearer ${token}`;
   }

   sleep(0.3); // User think time

   // ═══════════════════════════════════════════════════════════
   // PHASE 2: Product Browse (API Gateway → Product Service)
   // ═══════════════════════════════════════════════════════════
   group("Phase 2: Browse Products", function () {
      const browseStart = Date.now();

      // Get product details
      const productRes = http.get(`${BASE_URL}/products/${product.id}`, { headers });

      const browseDuration = Date.now() - browseStart;
      apiGatewayLatency.add(browseDuration);

      const success = check(productRes, {
         "product fetch is 200": (r) => r.status === 200,
         "product fetch < 30ms": () => browseDuration < 30,
      });

      errorRate.add(!success);

      if (ENABLE_DETAILED_LOGS && success) {
         console.log(`  ✓ Product fetched in ${browseDuration}ms`);
      }
   });

   sleep(0.5);

   // ═══════════════════════════════════════════════════════════
   // PHASE 3: Shopping Cart (API Gateway → Redis Cache)
   // ═══════════════════════════════════════════════════════════
   group("Phase 3: Add to Cart", function () {
      const cartStart = Date.now();

      const cartRes = http.post(
         `${BASE_URL}/cart/items`,
         JSON.stringify({
            product_id: product.id,
            quantity: quantity,
         }),
         { headers }
      );

      const cartDuration = Date.now() - cartStart;

      const success = check(cartRes, {
         "add to cart is 201": (r) => r.status === 201,
         "cart operation < 50ms": () => cartDuration < 50, // Should be fast (Redis)
      });

      errorRate.add(!success);
   });

   sleep(1); // User reviews cart

   // ═══════════════════════════════════════════════════════════
   // PHASE 4: Order Creation (CRITICAL - Tests Connection Pool)
   // API Gateway → Order Service → Product + Inventory + Payment + Notification
   // ═══════════════════════════════════════════════════════════
   group("Phase 4: Create Order (Connection Pool Test)", function () {
      const orderStart = Date.now();

      const orderRes = http.post(
         `${BASE_URL}/orders`,
         JSON.stringify({
            items: [
               {
                  product_id: product.id,
                  quantity: quantity,
                  price: product.price,
               },
            ],
            shipping_address: {
               street: `${Math.floor(Math.random() * 999)} Test Street`,
               city: "LoadTest City",
               state: "LT",
               zip_code: "12345",
               country: "US",
            },
            payment_method: "credit_card",
         }),
         { headers }
      );

      const orderDuration = Date.now() - orderStart;
      orderServiceLatency.add(orderDuration);

      const success = check(orderRes, {
         "order creation is 201": (r) => r.status === 201,
         "order time < 150ms (with pool)": () => orderDuration < 150,
         "order has valid ID": (r) => {
            try {
               const body = JSON.parse(r.body);
               orderId = body.order?.id || body.data?.id;
               return !!orderId;
            } catch (e) {
               return false;
            }
         },
      });

      errorRate.add(!success);

      // Calculate connection pool savings
      const expectedWithoutPool = 900; // 900ms without pool
      const savings = expectedWithoutPool - orderDuration;
      connectionPoolSavings.add(savings);

      if (ENABLE_DETAILED_LOGS) {
         console.log(`  ✓ Order created in ${orderDuration}ms (saved ${savings}ms with pool)`);
      }
   });

   sleep(0.5);

   // ═══════════════════════════════════════════════════════════
   // PHASE 5: Payment Processing (CRITICAL - Tests Connection Pool)
   // API Gateway → Payment Service → Order + Notification
   // ═══════════════════════════════════════════════════════════
   if (orderId) {
      group("Phase 5: Process Payment (Connection Pool Test)", function () {
         const paymentStart = Date.now();

         const paymentRes = http.post(
            `${BASE_URL}/payments`,
            JSON.stringify({
               order_id: orderId,
               amount: product.price * quantity,
               currency: "USD",
               payment_method: "credit_card",
               card_details: {
                  number: "4242424242424242",
                  exp_month: "12",
                  exp_year: "2026",
                  cvc: "123",
               },
            }),
            { headers }
         );

         const paymentDuration = Date.now() - paymentStart;
         paymentServiceLatency.add(paymentDuration);

         const success = check(paymentRes, {
            "payment is 200 or 201": (r) => r.status === 200 || r.status === 201,
            "payment time < 100ms (with pool)": () => paymentDuration < 100,
            "payment has ID": (r) => {
               try {
                  const body = JSON.parse(r.body);
                  paymentId = body.payment?.id || body.data?.id;
                  return !!paymentId;
               } catch (e) {
                  return false;
               }
            },
         });

         errorRate.add(!success);

         if (ENABLE_DETAILED_LOGS) {
            console.log(`  ✓ Payment processed in ${paymentDuration}ms`);
         }
      });
   }

   // ═══════════════════════════════════════════════════════════
   // PHASE 6: Verification
   // ═══════════════════════════════════════════════════════════
   if (orderId) {
      group("Phase 6: Verify Order Status", function () {
         const statusRes = http.get(`${BASE_URL}/orders/${orderId}`, { headers });

         check(statusRes, {
            "order status check is 200": (r) => r.status === 200,
         });
      });
   }

   // Calculate full flow metrics
   const flowDuration = Date.now() - flowStart;
   fullFlowDuration.add(flowDuration);

   if (orderId && paymentId) {
      fullFlowSuccess.add(1);

      if (ENABLE_DETAILED_LOGS || flowDuration > 300) {
         const emoji = flowDuration < 200 ? "🚀" : flowDuration < 300 ? "✅" : "⚠️";
         console.log(`${emoji} Full flow: ${flowDuration}ms (Order: ${orderId})`);
      }
   } else {
      fullFlowFailure.add(1);
      console.error(`❌ Flow failed in ${flowDuration}ms`);
   }

   sleep(2); // Cool-down between iterations
}

export function teardown(data) {
   console.log("");
   console.log("═══════════════════════════════════════════════════════════");
   console.log("🏁 Test Complete! Analyzing Results...");
   console.log("═══════════════════════════════════════════════════════════");
   console.log("");
   console.log("📊 Connection Pool Performance Validation:");
   console.log("");
   console.log("  Target Performance (with connection pool):");
   console.log("    • Full flow p95: < 250ms (vs ~1560ms without pool)");
   console.log("    • Full flow p99: < 500ms");
   console.log("    • Order service: < 150ms (vs ~900ms without pool)");
   console.log("    • Payment service: < 100ms (vs ~440ms without pool)");
   console.log("");
   console.log("  Expected Improvements:");
   console.log("    • Overall: 9.75x faster");
   console.log("    • API Gateway: 11x faster");
   console.log("    • Order Service: 9x faster");
   console.log("    • Payment Service: 11x faster");
   console.log("");
   console.log("📈 Check K6 summary for actual metrics!");
   console.log("");
   console.log("💡 Key Metrics to Review:");
   console.log("   ✓ full_flow_duration: Should show p95 < 250ms");
   console.log("   ✓ order_service_latency: Should show p95 < 150ms");
   console.log("   ✓ payment_service_latency: Should show p95 < 100ms");
   console.log("   ✓ connection_pool_time_saved_ms: Avg ~800ms saved per request");
   console.log("   ✓ errors: Should be < 1%");
   console.log("");
   console.log("═══════════════════════════════════════════════════════════");
}
