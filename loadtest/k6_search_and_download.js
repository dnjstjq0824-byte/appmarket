import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 100 }, // Warmup to 100 VUs
    { duration: '1m', target: 500 },  // Sustained load at 500 VUs
    { duration: '30s', target: 0 },    // Ramp down
  ],
  thresholds: {
    'http_req_duration{type:search}': ['p(95)<50'],   // Search p95 < 50ms
    'http_req_duration{type:download}': ['p(95)<30'], // Token issue p95 < 30ms
    'http_req_failed': ['rate<0.001'],                // Error rate < 0.1%
  },
};

const BASE_URL = __ENV.TARGET_URL || 'http://localhost:8080';

const KEYWORDS = [
  'office', 'editor', 'photo', 'camera', 'security',
  'vpn', 'music', 'game', 'rpg', 'wallet', 'crypto', 'notes'
];

export default function () {
  const keyword = KEYWORDS[Math.floor(Math.random() * KEYWORDS.length)];

  // 1. Search Query Test
  const searchRes = http.get(`${BASE_URL}/api/v1/search?q=${encodeURIComponent(keyword)}&limit=20`, {
    tags: { type: 'search' },
    headers: {
      'Accept': 'application/json',
      'X-Forwarded-For': `198.51.100.${Math.floor(Math.random() * 254) + 1}`,
    },
  });

  check(searchRes, {
    'search status is 200': (r) => r.status === 200,
    'search response time < 50ms': (r) => r.timings.duration < 50,
  });

  sleep(0.5);

  // 2. Tokenized Download URL Issuance Test
  const tokenPayload = JSON.stringify({
    app_id: 'app-super-office-01',
    package_name: 'com.omniwork.office',
    version_code: 102,
    device_hash: `dev-sim-${__VU}-${Math.floor(Math.random() * 1000)}`,
  });

  const tokenRes = http.post(`${BASE_URL}/api/v1/distribution/tokens`, tokenPayload, {
    tags: { type: 'download' },
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer simulated_user_jwt_token',
    },
  });

  check(tokenRes, {
    'token status is 200 or 201': (r) => r.status === 200 || r.status === 201,
    'token response time < 30ms': (r) => r.timings.duration < 30,
  });

  sleep(1);
}

