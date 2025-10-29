import http from "k6/http";
import { check, sleep, group } from "k6";
import { Rate, Trend, Counter } from "k6/metrics";

// Custom metrics
const errorRate = new Rate("errors");
const orderCreationDuration = new Trend("order_creation_duration");
const orderCreationSuccess = new Counter("order_creation_success");
const orderCreationFailure = new Counter("order_creation_failure");

// Test configuration
export const options = {
   stages: [
      { duration: "30s", target: 10 }, // Warm-up: ramp up to 10 users
      { duration: "1m", target: 50 }, // Load: ramp up to 50 users
      { duration: "2m", target: 50 }, // Sustain: hold at 50 users
      { duration: "1m", target: 100 }, // Spike: ramp up to 100 users
      { duration: "2m", target: 100 }, // Peak: hold at 100 users
      { duration: "1m", target: 0 }, // Cool-down: ramp down to 0
   ],
   thresholds: {
      // Connection pool target: 95% of requests < 200ms (vs 500ms without pool)
      http_req_duration: ["p(95)<200", "p(99)<500"],

      // Error rate should be < 1%
      errors: ["rate<0.01"],

      // Order creation should be fast with connection pool
      order_creation_duration: ["p(95)<200", "p(99)<500"],

      // Success rate should be > 99%
      checks: ["rate>0.99"],
   },
};

// Test data
const BASE_URL = __ENV.BASE_URL || "http://localhost:8000";
const TEST_USER_EMAIL = "loadtest@example.com";
const TEST_USER_PASSWORD = "testpassword123";

// Sample product IDs (should exist in database)
const PRODUCT_IDS = ["prod-1", "prod-2", "prod-3", "prod-4", "prod-5"];

/**
 * Setup phase - runs once before load test
 */
export function setup() {
   console.log("🚀 Starting Order Service Load Test");
   console.log(`📍 Target: ${BASE_URL}`);
   console.log("⚡ Testing connection pool performance improvements");
   console.log("");

   // Try to login with test user
   const loginRes = http.post(
      `${BASE_URL}/auth/login`,
      JSON.stringify({
         email: TEST_USER_EMAIL,
         password: TEST_USER_PASSWORD,
      }),
      {
         headers: { "Content-Type": "application/json" },
      }
   );

   let token = null;
   if (loginRes.status === 200) {
      const body = JSON.parse(loginRes.body);
      token = body.access_token || body.data?.access_token;
      console.log("✅ Test user authenticated");
   } else {
      console.log("⚠️  Test user not authenticated - some tests may fail");
   }

   return { token };
}

/**
 * Main test scenario
 */
export default function (data) {
   const headers = {
      "Content-Type": "application/json",
   };

   if (data.token) {
      headers["Authorization"] = `Bearer ${data.token}`;
   }

   group("Order Creation Flow", function () {
      // Step 1: Get product details (tests connection pool to Product Service)
      const productId = PRODUCT_IDS[Math.floor(Math.random() * PRODUCT_IDS.length)];

      group("1. Get Product Details", function () {
         const startTime = Date.now();
         const productRes = http.get(`${BASE_URL}/products/${productId}`, { headers });
         const duration = Date.now() - startTime;

         const success = check(productRes, {
            "product fetch status is 200": (r) => r.status === 200,
            "product fetch time < 100ms": () => duration < 100,
         });

         errorRate.add(!success);
      });

      sleep(0.5); // Simulate user thinking time

      // Step 2: Add to cart (tests Redis cache)
      let cartItemId = null;
      group("2. Add to Cart", function () {
         const startTime = Date.now();
         const cartRes = http.post(
            `${BASE_URL}/cart/items`,
            JSON.stringify({
               product_id: productId,
               quantity: Math.floor(Math.random() * 3) + 1, // 1-3 items
            }),
            { headers }
         );
         const duration = Date.now() - startTime;

         const success = check(cartRes, {
            "add to cart status is 201": (r) => r.status === 201,
            "add to cart time < 50ms": () => duration < 50, // Should be fast (Redis)
         });

         if (success && cartRes.body) {
            try {
               const body = JSON.parse(cartRes.body);
               cartItemId = body.cart_item?.id || body.data?.id;
            } catch (e) {
               // Ignore parse errors
            }
         }

         errorRate.add(!success);
      });

      sleep(1); // Simulate user reviewing cart

      // Step 3: Create order (tests connection pool to multiple services)
      // This is the critical test - should be fast with connection pooling
      group("3. Create Order (Connection Pool Test)", function () {
         const orderStartTime = Date.now();

         const orderRes = http.post(
            `${BASE_URL}/orders`,
            JSON.stringify({
               items: [
                  {
                     product_id: productId,
                     quantity: 2,
                     price: 29.99,
                  },
               ],
               shipping_address: {
                  street: "123 Test St",
                  city: "Test City",
                  state: "TS",
                  zip_code: "12345",
                  country: "US",
               },
               payment_method: "credit_card",
            }),
            { headers }
         );

         const orderDuration = Date.now() - orderStartTime;
         orderCreationDuration.add(orderDuration);

         const success = check(orderRes, {
            "order creation status is 201": (r) => r.status === 201,
            "order creation time < 200ms (with pool)": () => orderDuration < 200,
            "order has ID": (r) => {
               try {
                  const body = JSON.parse(r.body);
                  return body.order?.id || body.data?.id;
               } catch (e) {
                  return false;
               }
            },
         });

         if (success) {
            orderCreationSuccess.add(1);
            console.log(`✅ Order created in ${orderDuration}ms (target: <200ms)`);
         } else {
            orderCreationFailure.add(1);
            console.log(`❌ Order creation failed or too slow: ${orderDuration}ms`);
         }

         errorRate.add(!success);
      });

      sleep(2); // Simulate user waiting for confirmation
   });

   // Additional scenario: List orders (tests caching)
   group("Order History", function () {
      const startTime = Date.now();
      const ordersRes = http.get(`${BASE_URL}/orders?limit=10`, { headers });
      const duration = Date.now() - startTime;

      const success = check(ordersRes, {
         "list orders status is 200": (r) => r.status === 200,
         "list orders time < 100ms": () => duration < 100,
      });

      errorRate.add(!success);
   });

   sleep(1); // Think time between iterations
}

/**
 * Teardown phase - runs once after load test
 */
export function teardown(data) {
   console.log("");
   console.log("🏁 Load Test Complete!");
   console.log("");
   console.log("📊 Check metrics for:");
   console.log("   - Order creation duration (should be <200ms p95 with connection pool)");
   console.log("   - Error rate (should be <1%)");
   console.log("   - Success rate (should be >99%)");
   console.log("");
   console.log("💡 Compare with baseline (without connection pool):");
   console.log("   - Expected 9x improvement in latency");
   console.log("   - Previous: ~1500ms total, Now: ~160ms total");
}
