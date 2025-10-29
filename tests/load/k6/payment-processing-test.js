import http from "k6/http";
import { check, sleep, group } from "k6";
import { Rate, Trend, Counter } from "k6/metrics";

// Custom metrics
const errorRate = new Rate("errors");
const paymentProcessingDuration = new Trend("payment_processing_duration");
const paymentSuccess = new Counter("payment_success");
const paymentFailure = new Counter("payment_failure");
const endToEndDuration = new Trend("end_to_end_duration"); // Order + Payment

// Test configuration
export const options = {
   stages: [
      { duration: "30s", target: 10 }, // Warm-up
      { duration: "1m", target: 25 }, // Load
      { duration: "2m", target: 25 }, // Sustain
      { duration: "1m", target: 50 }, // Spike
      { duration: "2m", target: 50 }, // Peak
      { duration: "1m", target: 0 }, // Cool-down
   ],
   thresholds: {
      // Payment should be FAST with connection pool
      http_req_duration: ["p(95)<150", "p(99)<300"],

      // Payment-specific thresholds
      payment_processing_duration: ["p(95)<100", "p(99)<200"],

      // End-to-end: Order + Payment (target: <200ms vs 1560ms before)
      end_to_end_duration: ["p(95)<250", "p(99)<500"],

      // Error rate
      errors: ["rate<0.01"],

      // Success rate
      checks: ["rate>0.99"],
   },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8000";
const TEST_USER_EMAIL = "loadtest@example.com";
const TEST_USER_PASSWORD = "testpassword123";

export function setup() {
   console.log("🚀 Starting Payment Service Load Test");
   console.log(`📍 Target: ${BASE_URL}`);
   console.log("⚡ Testing connection pool: Payment Service → Order + Notification");
   console.log("");

   // Authenticate
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
   }

   return { token };
}

export default function (data) {
   const headers = {
      "Content-Type": "application/json",
   };

   if (data.token) {
      headers["Authorization"] = `Bearer ${data.token}`;
   }

   let orderId = null;
   const flowStartTime = Date.now();

   // Step 1: Create an order first
   group("1. Create Order", function () {
      const orderRes = http.post(
         `${BASE_URL}/orders`,
         JSON.stringify({
            items: [
               {
                  product_id: "prod-1",
                  quantity: 1,
                  price: 49.99,
               },
            ],
            shipping_address: {
               street: "456 Payment Test Ave",
               city: "Test City",
               state: "TS",
               zip_code: "54321",
               country: "US",
            },
            payment_method: "credit_card",
         }),
         { headers }
      );

      const success = check(orderRes, {
         "order created": (r) => r.status === 201,
      });

      if (success && orderRes.body) {
         try {
            const body = JSON.parse(orderRes.body);
            orderId = body.order?.id || body.data?.id;
         } catch (e) {
            console.error("Failed to parse order response");
         }
      }

      errorRate.add(!success);
   });

   sleep(0.5);

   // Step 2: Process payment (CRITICAL - tests Payment Service connection pool)
   if (orderId) {
      group("2. Process Payment (Connection Pool Test)", function () {
         const paymentStartTime = Date.now();

         const paymentRes = http.post(
            `${BASE_URL}/payments`,
            JSON.stringify({
               order_id: orderId,
               amount: 49.99,
               currency: "USD",
               payment_method: "credit_card",
               card_details: {
                  number: "4242424242424242",
                  exp_month: "12",
                  exp_year: "2025",
                  cvc: "123",
               },
            }),
            { headers }
         );

         const paymentDuration = Date.now() - paymentStartTime;
         const totalDuration = Date.now() - flowStartTime;

         paymentProcessingDuration.add(paymentDuration);
         endToEndDuration.add(totalDuration);

         const success = check(paymentRes, {
            "payment status is 201 or 200": (r) => r.status === 201 || r.status === 200,
            "payment time < 100ms (with pool)": () => paymentDuration < 100,
            "end-to-end < 250ms (order + payment)": () => totalDuration < 250,
            "payment has ID": (r) => {
               try {
                  const body = JSON.parse(r.body);
                  return body.payment?.id || body.data?.id;
               } catch (e) {
                  return false;
               }
            },
         });

         if (success) {
            paymentSuccess.add(1);
            console.log(`✅ Payment processed in ${paymentDuration}ms (E2E: ${totalDuration}ms)`);
         } else {
            paymentFailure.add(1);
            console.log(`❌ Payment failed: ${paymentDuration}ms (E2E: ${totalDuration}ms)`);
         }

         errorRate.add(!success);
      });
   } else {
      console.error("❌ Skipping payment - no order ID");
   }

   sleep(1);

   // Step 3: Verify order status was updated (tests Order Service call from Payment)
   if (orderId) {
      group("3. Verify Order Status", function () {
         const statusRes = http.get(`${BASE_URL}/orders/${orderId}`, { headers });

         const success = check(statusRes, {
            "order status check is 200": (r) => r.status === 200,
            "order status updated": (r) => {
               try {
                  const body = JSON.parse(r.body);
                  const status = body.order?.status || body.data?.status;
                  return status === "PAID" || status === "PROCESSING" || status === "CONFIRMED";
               } catch (e) {
                  return false;
               }
            },
         });

         errorRate.add(!success);
      });
   }

   sleep(2);
}

export function teardown(data) {
   console.log("");
   console.log("🏁 Payment Load Test Complete!");
   console.log("");
   console.log("📊 Key Metrics:");
   console.log("   - Payment processing: Should be <100ms p95 (vs ~440ms without pool)");
   console.log("   - End-to-end (Order + Payment): Should be <250ms p95 (vs ~1560ms without pool)");
   console.log("   - Expected improvement: 11x faster for payment, 6-9x faster E2E");
   console.log("");
   console.log("🔍 Connection Pool Impact:");
   console.log("   - Payment → Order Service: 220ms → 20ms (11x)");
   console.log("   - Payment → Notification: 220ms → 20ms (11x)");
   console.log("   - Total payment overhead: 440ms → 40ms (11x)");
}
