import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Counter, Trend } from 'k6/metrics';
import { SharedArray } from 'k6/data';

// ============================================================
// CONFIG - Sesuaikan sebelum run
// ============================================================
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const AUTH_EMAIL = __ENV.AUTH_EMAIL || 'admin@swiftlead.com';
const AUTH_PASSWORD = __ENV.AUTH_PASSWORD || 'password123';

// ============================================================
// CUSTOM METRICS
// ============================================================
const errorRate = new Rate('errors');
const loginFailures = new Counter('login_failures');
const apiLatency = new Trend('api_latency', true);

// ============================================================
// STAGES - Stress test brutal: ramp up sampai server kewalahan
// ============================================================
export const options = {
  scenarios: {
    // Scenario 1: Spike test - langsung hajar 500 VU
    spike: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '10s', target: 50 },    // warm up
        { duration: '20s', target: 200 },   // ramp up
        { duration: '30s', target: 500 },   // heavy load
        { duration: '20s', target: 1000 },  // brutal spike
        { duration: '60s', target: 1000 },  // sustain brutal load
        { duration: '30s', target: 1500 },  // push to breaking point
        { duration: '60s', target: 1500 },  // sustain breaking point
        { duration: '20s', target: 2000 },  // absolute chaos
        { duration: '60s', target: 2000 },  // hold chaos
        { duration: '30s', target: 0 },     // cool down
      ],
      gracefulRampDown: '10s',
    },

    // Scenario 2: Constant high-rate requests (fire-and-forget)
    firehose: {
      executor: 'constant-arrival-rate',
      rate: 500,              // 500 requests per second
      timeUnit: '1s',
      duration: '5m',
      preAllocatedVUs: 200,
      maxVUs: 1000,
      startTime: '30s',       // start after spike warms up
    },
  },

  thresholds: {
    http_req_duration: ['p(95)<2000'],  // akan fail, tapi kita mau lihat seberapa parah
    errors: ['rate<0.5'],               // expect banyak error
  },
};

// ============================================================
// SETUP - Login dan ambil token
// ============================================================
export function setup() {
  const loginRes = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify({
    email: AUTH_EMAIL,
    password: AUTH_PASSWORD,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  if (loginRes.status !== 200) {
    console.error(`Login failed: ${loginRes.status} - ${loginRes.body}`);
    return { token: '' };
  }

  const body = JSON.parse(loginRes.body);
  const token = body.data?.token || body.token || '';
  console.log(`Login successful, token obtained`);

  // Get some IDs for testing
  const headers = authHeaders(token);

  const rbwRes = http.get(`${BASE_URL}/api/v1/rbw`, { headers });
  let rbwIds = [];
  if (rbwRes.status === 200) {
    const rbwData = JSON.parse(rbwRes.body);
    rbwIds = (rbwData.data || []).map(r => r.id).slice(0, 5);
  }

  const nodeRes = rbwIds.length > 0
    ? http.get(`${BASE_URL}/api/v1/rbw/${rbwIds[0]}/nodes`, { headers })
    : null;
  let nodeIds = [];
  if (nodeRes && nodeRes.status === 200) {
    const nodeData = JSON.parse(nodeRes.body);
    nodeIds = (nodeData.data || []).map(n => n.id).slice(0, 5);
  }

  return { token, rbwIds, nodeIds };
}

function authHeaders(token) {
  return {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };
}

// ============================================================
// MAIN TEST FUNCTION
// ============================================================
export default function (data) {
  const headers = authHeaders(data.token);
  const rbwIds = data.rbwIds || [];
  const nodeIds = data.nodeIds || [];

  // Randomly pick which endpoint group to hit
  const scenario = Math.random();

  if (scenario < 0.15) {
    hammmerHealth();
  } else if (scenario < 0.30) {
    hammerAuth(data);
  } else if (scenario < 0.50) {
    hammerRBW(headers, rbwIds);
  } else if (scenario < 0.65) {
    hammerNodes(headers, nodeIds);
  } else if (scenario < 0.80) {
    hammerSensors(headers, nodeIds);
  } else if (scenario < 0.90) {
    hammerAlerts(headers, rbwIds);
  } else {
    hammerAI(headers, nodeIds);
  }

  // Minimal sleep - kita mau brutal
  sleep(Math.random() * 0.1);
}

// ============================================================
// ATTACK FUNCTIONS
// ============================================================

function hammmerHealth() {
  group('Health Check Flood', function () {
    // Spam health check - simple tapi banyak
    for (let i = 0; i < 10; i++) {
      const res = http.get(`${BASE_URL}/health`);
      errorRate.add(res.status !== 200);
      apiLatency.add(res.timings.duration);
    }
  });
}

function hammerAuth(data) {
  group('Auth Bruteforce', function () {
    // Spam login attempts (valid + invalid)
    for (let i = 0; i < 5; i++) {
      const payload = JSON.stringify({
        email: Math.random() > 0.3 ? AUTH_EMAIL : `fake${__VU}_${i}@test.com`,
        password: Math.random() > 0.3 ? AUTH_PASSWORD : 'wrongpassword',
      });

      const res = http.post(`${BASE_URL}/api/v1/auth/login`, payload, {
        headers: { 'Content-Type': 'application/json' },
      });

      if (res.status !== 200) loginFailures.add(1);
      errorRate.add(res.status >= 500);
      apiLatency.add(res.timings.duration);
    }

    // Spam register attempts
    const regPayload = JSON.stringify({
      name: `StressUser_${__VU}_${Date.now()}`,
      email: `stress_${__VU}_${Date.now()}@loadtest.com`,
      password: 'LoadTest123!',
    });
    const res = http.post(`${BASE_URL}/api/v1/auth/register`, regPayload, {
      headers: { 'Content-Type': 'application/json' },
    });
    errorRate.add(res.status >= 500);
    apiLatency.add(res.timings.duration);
  });
}

function hammerRBW(headers, rbwIds) {
  group('RBW CRUD Storm', function () {
    // List RBW - heavy query
    for (let i = 0; i < 5; i++) {
      const res = http.get(`${BASE_URL}/api/v1/rbw`, { headers });
      errorRate.add(res.status >= 500);
      apiLatency.add(res.timings.duration);
    }

    // Get individual RBW
    if (rbwIds.length > 0) {
      const id = rbwIds[Math.floor(Math.random() * rbwIds.length)];
      for (let i = 0; i < 3; i++) {
        const res = http.get(`${BASE_URL}/api/v1/rbw/${id}`, { headers });
        errorRate.add(res.status >= 500);
        apiLatency.add(res.timings.duration);
      }

      // Get nested resources
      http.get(`${BASE_URL}/api/v1/rbw/${id}/nodes`, { headers });
      http.get(`${BASE_URL}/api/v1/rbw/${id}/alerts`, { headers });
      http.get(`${BASE_URL}/api/v1/rbw/${id}/harvests`, { headers });
      http.get(`${BASE_URL}/api/v1/rbw/${id}/transactions`, { headers });
    }

    // Create RBW (write pressure)
    const createPayload = JSON.stringify({
      name: `LoadTest_RBW_${__VU}_${Date.now()}`,
      location: `Stress Test Location ${__VU}`,
    });
    const res = http.post(`${BASE_URL}/api/v1/rbw`, createPayload, { headers });
    errorRate.add(res.status >= 500);
    apiLatency.add(res.timings.duration);
  });
}

function hammerNodes(headers, nodeIds) {
  group('Node Operations Flood', function () {
    if (nodeIds.length === 0) {
      // Fallback: just hit random UUIDs to generate 404s/errors
      for (let i = 0; i < 10; i++) {
        const fakeId = `00000000-0000-0000-0000-${String(i).padStart(12, '0')}`;
        http.get(`${BASE_URL}/api/v1/nodes/${fakeId}`, { headers });
        http.get(`${BASE_URL}/api/v1/nodes/${fakeId}/sensors`, { headers });
      }
      return;
    }

    for (const nodeId of nodeIds) {
      // Get node
      const res1 = http.get(`${BASE_URL}/api/v1/nodes/${nodeId}`, { headers });
      errorRate.add(res1.status >= 500);
      apiLatency.add(res1.timings.duration);

      // Get sensors
      const res2 = http.get(`${BASE_URL}/api/v1/nodes/${nodeId}/sensors`, { headers });
      errorRate.add(res2.status >= 500);
      apiLatency.add(res2.timings.duration);

      // Get audio state
      http.get(`${BASE_URL}/api/v1/nodes/${nodeId}/audio`, { headers });

      // Toggle pump mode (write operation - dangerous under load)
      http.patch(`${BASE_URL}/api/v1/nodes/${nodeId}/pump/mode`, JSON.stringify({
        mode: Math.random() > 0.5 ? 'auto' : 'manual',
      }), { headers });
    }
  });
}

function hammerSensors(headers, nodeIds) {
  group('Sensor Data Flood', function () {
    if (nodeIds.length === 0) return;

    const nodeId = nodeIds[Math.floor(Math.random() * nodeIds.length)];

    // Get sensors for node
    const sensorsRes = http.get(`${BASE_URL}/api/v1/nodes/${nodeId}/sensors`, { headers });
    if (sensorsRes.status !== 200) return;

    let sensors = [];
    try {
      const data = JSON.parse(sensorsRes.body);
      sensors = data.data || [];
    } catch (e) {
      return;
    }

    // Hammer each sensor with readings requests
    for (const sensor of sensors.slice(0, 3)) {
      for (let i = 0; i < 5; i++) {
        // Get readings (heavy DB query)
        const res = http.get(`${BASE_URL}/api/v1/sensors/${sensor.id}/readings`, { headers });
        errorRate.add(res.status >= 500);
        apiLatency.add(res.timings.duration);

        // Get trend (another heavy query)
        http.get(`${BASE_URL}/api/v1/sensors/${sensor.id}/trend`, { headers });

        // Create fake reading (write flood)
        const readingPayload = JSON.stringify({
          value: Math.random() * 100,
          recorded_at: new Date().toISOString(),
        });
        http.post(`${BASE_URL}/api/v1/sensors/${sensor.id}/readings`, readingPayload, { headers });
      }
    }
  });
}

function hammerAlerts(headers, rbwIds) {
  group('Alert System Stress', function () {
    // List all alerts
    for (let i = 0; i < 10; i++) {
      const res = http.get(`${BASE_URL}/api/v1/alerts`, { headers });
      errorRate.add(res.status >= 500);
      apiLatency.add(res.timings.duration);
    }

    // Get alerts per RBW
    for (const rbwId of rbwIds) {
      http.get(`${BASE_URL}/api/v1/rbw/${rbwId}/alerts`, { headers });
    }
  });
}

function hammerAI(headers, nodeIds) {
  group('AI Engine Overload', function () {
    if (nodeIds.length === 0) return;

    const nodeId = nodeIds[Math.floor(Math.random() * nodeIds.length)];

    // AI endpoints are expensive - spam them
    for (let i = 0; i < 3; i++) {
      const res1 = http.post(`${BASE_URL}/api/v1/nodes/${nodeId}/ai/analyze`, null, { headers });
      errorRate.add(res1.status >= 500);
      apiLatency.add(res1.timings.duration);

      http.post(`${BASE_URL}/api/v1/nodes/${nodeId}/ai/predict-pump`, null, { headers });
      http.post(`${BASE_URL}/api/v1/nodes/${nodeId}/ai/predict-grade`, null, { headers });
      http.post(`${BASE_URL}/api/v1/nodes/${nodeId}/ai/anomaly-detect`, null, { headers });
    }

    // Also hit the general AI endpoints
    http.get(`${BASE_URL}/api/v1/ai/health`, { headers });
    http.post(`${BASE_URL}/api/v1/ai/analyze`, JSON.stringify({
      temperature: 35 + Math.random() * 10,
      humidity: 60 + Math.random() * 30,
      ammonia: Math.random() * 50,
    }), { headers });
  });
}

// ============================================================
// TEARDOWN
// ============================================================
export function teardown(data) {
  console.log('=== STRESS TEST COMPLETE ===');
  console.log('Check Grafana dashboard for CPU/memory impact');
}
