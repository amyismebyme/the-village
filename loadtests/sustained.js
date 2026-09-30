import { sleep } from 'k6';

import {
  getHealth,
  getReady,
  getStatus,
  listCommunities,
} from './lib/workload.js';

const p95LimitMs = Number(__ENV.K6_P95_MS || '2000');
const errorRateLimit = __ENV.K6_ERROR_RATE || '0.02';

export const options = {
  stages: [
    { duration: '1m', target: 10 },
    { duration: '2m', target: 25 },
    { duration: '3m', target: 50 },
    { duration: '2m', target: 25 },
    { duration: '1m', target: 10 },
    { duration: '1m', target: 0 },
  ],
  thresholds: {
    http_req_failed: [`rate<${errorRateLimit}`],
    http_req_duration: [`p(95)<${p95LimitMs}`],
  },
};

export default function () {
  const roll = Math.random();

  if (roll < 0.06) {
    getHealth();
  } else if (roll < 0.12) {
    getReady();
  } else if (roll < 0.17) {
    getStatus();
  } else {
    listCommunities(20, 0);
  }

  sleep(0.15);
}
