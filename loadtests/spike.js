import { sleep } from 'k6';

import {
  getHealth,
  getReady,
  getStatus,
  listCommunities,
} from './lib/workload.js';

const p95LimitMs = Number(__ENV.K6_P95_MS || '3000');
const errorRateLimit = __ENV.K6_ERROR_RATE || '0.05';

export const options = {
  stages: [
    { duration: '30s', target: 5 },
    { duration: '10s', target: 75 },
    { duration: '60s', target: 75 },
    { duration: '10s', target: 5 },
    { duration: '60s', target: 5 },
    { duration: '20s', target: 0 },
  ],
  thresholds: {
    http_req_failed: [`rate<${errorRateLimit}`],
    http_req_duration: [`p(95)<${p95LimitMs}`],
  },
};

export default function () {
  const roll = Math.random();

  if (roll < 0.04) {
    getHealth();
  } else if (roll < 0.08) {
    getReady();
  } else if (roll < 0.12) {
    getStatus();
  } else {
    listCommunities(20, 0);
  }

  sleep(0.10);
}
