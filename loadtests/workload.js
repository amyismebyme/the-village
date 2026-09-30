import { check } from 'k6';
import {
  getHealth,
  getReady,
  getCommunities,
  getCommunity,
} from './common.js';

export function exerciseReadWorkload(data) {
  const roll = Math.random();

  if (roll < 0.10) {
    const response = getHealth();

    check(response, {
      'health returns 200': (r) => r.status === 200,
    });

    return;
  }

  if (roll < 0.25) {
    const response = getReady();

    check(response, {
      'ready returns 200': (r) => r.status === 200,
    });

    return;
  }

  if (roll < 0.60) {
    const response = getCommunities();

    check(response, {
      'community list returns 200': (r) => r.status === 200,
      'community list returns JSON': (r) =>
        String(r.headers['Content-Type'] || '').includes('application/json'),
    });

    return;
  }

  const ids = data.ids || [];
  const id = ids[Math.floor(Math.random() * ids.length)];

  const response = getCommunity(id);

  check(response, {
    'community get returns 200': (r) => r.status === 200,
    'community get returns JSON': (r) =>
      String(r.headers['Content-Type'] || '').includes('application/json'),
  });
}
