import http from 'k6/http';
import { check, sleep } from 'k6';

// 10,000 QPS Enterprise Stress & Concurrency Benchmark
export const options = {
  scenarios: {
    constant_request_rate: {
      executor: 'constant-arrival-rate',
      rate: 10000, // 10,000 requests per second
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 1000,
      maxVUs: 4000,
    },
  },
  thresholds: {
    'http_req_duration{type:health}': ['p(95)<15', 'p(99)<30'],
    'http_req_duration{type:catalog}': ['p(95)<35', 'p(99)<70'],
    'http_req_duration{type:distribution}': ['p(95)<40', 'p(99)<80'],
    'http_req_failed': ['rate<0.001'], // 99.99% availability
  },
};

const BASE_URL = __ENV.TARGET_URL || 'https://appmarket.dnjstjqppk.workers.dev';

export default function () {
  const rand = Math.random();

  if (rand < 0.2) {
    // 20% Healthcheck / keepalive traffic
    const res = http.get(`${BASE_URL}/healthz/ready`, {
      tags: { type: 'health' },
    });
    check(res, {
      'health status 200': (r) => r.status === 200,
    });
  } else if (rand < 0.7) {
    // 50% Catalog Discovery & Search Traffic
    const categories = ['ENTERPRISE_ERP', 'SECURITY_PRIVACY', 'FINTECH_DEFI', 'DEV_TOOLS', 'GAMES'];
    const cat = categories[Math.floor(Math.random() * categories.length)];
    const res = http.get(`${BASE_URL}/api/v1/catalog/apps?category=${cat}&page=1&limit=20`, {
      tags: { type: 'catalog' },
      headers: {
        'Accept': 'application/json',
        'X-Edge-Routing': 'geo-optimized',
      },
    });
    check(res, {
      'catalog status 200': (r) => r.status === 200,
    });
  } else {
    // 30% Distribution token issuance
    const payload = JSON.stringify({
      app_id: 'app-omni-office-pro',
      package_name: 'com.omniwork.office',
      version_code: 102,
      device_hash: `dev-node-${__VU}-${Date.now()}`,
    });
    const res = http.post(`${BASE_URL}/api/v1/distribution/tokens`, payload, {
      tags: { type: 'distribution' },
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer simulated-enterprise-token',
      },
    });
    check(res, {
      'token status 200 or 201': (r) => r.status === 200 || r.status === 201,
    });
  }
}

