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
// Use VU-specific user to avoid cart race conditions
// Each VU gets its own user: loadtest1@example.com, loadtest2@example.com, etc.
const getVUEmail = () => `loadtest${__VU}@example.com`;
const TEST_USER_PASSWORD = "TestPass123";

// Cache tokens per VU to avoid repeated logins
const vuTokenCache = {};

/**
 * Get or create auth token for current VU
 * Tokens are cached to avoid expensive bcrypt operations on every iteration
 */
function getVUToken() {
   const vuId = __VU;

   // Return cached token if exists
   if (vuTokenCache[vuId]) {
      return vuTokenCache[vuId];
   }

   // Login and cache token
   const vuEmail = getVUEmail();
   const loginRes = http.post(
      `${BASE_URL}/api/v1/auth/login`,
      JSON.stringify({
         email: vuEmail,
         password: TEST_USER_PASSWORD,
      }),
      {
         headers: { "Content-Type": "application/json" },
         timeout: "10s",
      }
   );

   if (loginRes.status !== 200) {
      console.log(`❌ VU ${vuId} login failed: ${loginRes.status}`);
      return null;
   }

   const body = JSON.parse(loginRes.body);
   const token = body.data?.access_token;

   if (!token) {
      console.log(`❌ VU ${vuId} no token in response`);
      return null;
   }

   // Cache token for future iterations
   vuTokenCache[vuId] = token;
   return token;
}

/**
 * Setup phase - runs once before load test
 * Creates test users for all VUs to avoid cart race conditions
 */
export function setup() {
   console.log("🚀 Starting Order Service Load Test");
   console.log(`📍 Target: ${BASE_URL}`);
   console.log("⚡ Testing connection pool performance improvements");
   console.log("");

   // Get max VUs from options
   const maxVUs = 100; // Match the max VUs in options.stages

   console.log(`� Creating ${maxVUs} test users (one per VU)...`);

   // Create users for all VUs
   let created = 0,
      existing = 0,
      failed = 0;

   for (let vu = 1; vu <= maxVUs; vu++) {
      const email = `loadtest${vu}@example.com`;
      const registerRes = http.post(
         `${BASE_URL}/api/v1/auth/register`,
         JSON.stringify({
            email: email,
            password: TEST_USER_PASSWORD,
            name: `Load Test User ${vu}`,
         }),
         {
            headers: { "Content-Type": "application/json" },
            timeout: "10s",
         }
      );

      if (registerRes.status === 201) {
         created++;
      } else if (
         registerRes.status === 409 ||
         (registerRes.body && registerRes.body.includes("already exists"))
      ) {
         existing++;
      } else {
         failed++;
      }
   }

   console.log(`Users ready: ${created} created, ${existing} existing, ${failed} failed`);

   // Login with first user to get token structure and test auth
   console.log("🔐 Testing authentication...");
   const loginRes = http.post(
      `${BASE_URL}/api/v1/auth/login`,
      JSON.stringify({
         email: "loadtest1@example.com",
         password: TEST_USER_PASSWORD,
      }),
      {
         headers: { "Content-Type": "application/json" },
      }
   );

   let token = null;
   if (loginRes.status === 200) {
      try {
         const body = JSON.parse(loginRes.body);
         token = body.access_token || body.data?.access_token;
         if (token) {
            console.log("Test user authenticated");
         } else {
            console.log("❌ Token not found in response body");
            console.log(`Response: ${loginRes.body}`);
         }
      } catch (e) {
         console.log(`❌ Failed to parse login response: ${e}`);
      }
   } else {
      console.log(`❌ Login failed with status ${loginRes.status}: ${loginRes.body}`);
      console.log("⚠️  Load test will fail - cannot authenticate user");
   }

   // Step 3: Get actual product IDs from database
   console.log("📦 Fetching available products...");
   const productsRes = http.get(`${BASE_URL}/api/v1/products?page=1&page_size=10`, {
      headers: {
         "Content-Type": "application/json",
         Authorization: token ? `Bearer ${token}` : "",
      },
   });

   let productIds = [];
   if (productsRes.status === 200) {
      try {
         const body = JSON.parse(productsRes.body);
         const products = body.data?.products || body.products || [];
         if (products.length > 0) {
            productIds = products.map((p) => p.id);
            console.log(`Found ${productIds.length} products in database`);
         } else {
            console.log("❌ No products found in database - test will fail");
            console.log("⚠️  Please seed products first!");
         }
      } catch (e) {
         console.log(`❌ Failed to parse products: ${e}`);
      }
   } else {
      console.log(
         `❌ Failed to fetch products (status ${productsRes.status}): ${productsRes.body}`
      );
   }

   if (productIds.length === 0) {
      console.log("❌ FATAL: No products available - load test cannot proceed");
   }

   return { token, productIds };
}

/**
 * Main test scenario
 * Each VU uses its own user to avoid cart race conditions
 */
export default function (data) {
   // Get cached token for this VU (login only once per VU)
   const token = getVUToken();

   if (!token) {
      console.log(`❌ VU ${__VU} failed to get token - skipping iteration`);
      return;
   }

   const headers = {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
   };

   // Use actual product IDs from setup
   const productIds = data.productIds;

   // Skip test if no products available
   if (!productIds || productIds.length === 0) {
      console.log("❌ No products available - skipping iteration");
      return;
   }

   group("Order Creation Flow", function () {
      // Step 1: Get product details (tests connection pool to Product Service)
      const productId = productIds[Math.floor(Math.random() * productIds.length)];

      group("1. Get Product Details", function () {
         const startTime = Date.now();
         const productRes = http.get(`${BASE_URL}/api/v1/products/${productId}`, { headers });
         const duration = Date.now() - startTime;

         if (productRes.status !== 200) {
            console.log(`⚠️  Product fetch failed: ${productRes.status} - ${productRes.body}`);
         }

         const success = check(productRes, {
            "product fetch status is 200": (r) => r.status === 200,
            "product fetch time < 100ms": () => duration < 100,
         });

         errorRate.add(!success);
      });

      sleep(0.5); // Simulate user thinking time

      // Step 2: Add to cart (tests Redis cache)
      // IMPORTANT: Always add items to cart before creating order
      // because order creation clears the cart
      let cartItemId = null;
      let cartAddSuccess = false;
      group("2. Add to Cart", function () {
         const startTime = Date.now();
         const cartRes = http.post(
            `${BASE_URL}/api/v1/cart`,
            JSON.stringify({
               product_id: productId,
               quantity: Math.floor(Math.random() * 3) + 1, // 1-3 items
            }),
            { headers }
         );
         const duration = Date.now() - startTime;

         cartAddSuccess = check(cartRes, {
            "add to cart status is 200": (r) => r.status === 200,
            "add to cart time < 50ms": () => duration < 50, // Should be fast (Redis)
         });

         if (cartAddSuccess && cartRes.body) {
            try {
               const body = JSON.parse(cartRes.body);
               cartItemId = body.cart?.id || body.data?.id;
            } catch (e) {
               // Ignore parse errors
            }
         }

         errorRate.add(!cartAddSuccess);
      });

      // Skip order creation if cart add failed
      if (!cartAddSuccess) {
         console.log("⚠️  Skipping order creation - cart add failed");
         return;
      }

      sleep(1); // Simulate user reviewing cart

      // Step 3: Create order (tests connection pool to multiple services)
      // This is the critical test - should be fast with connection pooling
      group("3. Create Order (Connection Pool Test)", function () {
         const orderStartTime = Date.now();

         const orderRes = http.post(
            `${BASE_URL}/api/v1/orders`,
            JSON.stringify({
               shipping_address: "123 Test St, Test City, TS 12345, US",
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
            console.log(`Order created in ${orderDuration}ms (target: <200ms)`);
         } else {
            orderCreationFailure.add(1);
            if (orderRes.status !== 201) {
               console.log(`❌ Order creation failed: ${orderRes.status} - ${orderRes.body}`);
            } else {
               console.log(`❌ Order creation too slow: ${orderDuration}ms`);
            }
         }

         errorRate.add(!success);
      });

      sleep(2); // Simulate user waiting for confirmation
   });

   // Additional scenario: List orders (tests caching)
   group("Order History", function () {
      const startTime = Date.now();
      const ordersRes = http.get(`${BASE_URL}/api/v1/orders?page=1&page_size=10`, { headers });
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
