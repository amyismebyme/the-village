import { sleep } from 'k6';

import {
  getHealth,
  getReady,
  getStatus,
  listCommunities,
} from './lib/workload.js';

const p95LimitMs = Number(__ENV.K6_P95_MS || '1000');
const errorRateLimit = __ENV.K6_ERROR_RATE || '0.01';

export const options = {
  vus: Number(__ENV.K6_VUS || '10'),
  duration: __ENV.K6_DURATION || '60s',
  thresholds: {
    http_req_failed: [`rate<${errorRateLimit}`],
    http_req_duration: [`p(95)<${p95LimitMs}`],
  },
};

export default function () {
  const roll = Math.random();

  if (roll < 0.08) {
    getHealth();
  } else if (roll < 0.16) {
    getReady();
  } else if (roll < 0.21) {
    getStatus();
  } else {
    listCommunities(20, 0);
  }

  sleep(0.2);
}
