
import {
  createCommunity,
  deleteCommunity,
  getHealth,
  getReady,
  listCommunities,
  updateCommunity,
} from './lib/workload.js';
import { uniqueSlug } from './lib/config.js';

export const options = {
  vus: 1,
  iterations: 1,
  thresholds: {
    http_req_failed: ['rate<0.02'],
    http_req_duration: ['p(95)<1500'],
  },
};

export default function () {
  getHealth();
  getReady();

  listCommunities();

  const slug = uniqueSlug('k6-smoke');
  const create = createCommunity('K6 Performance Smoke Test', slug);

  let id = null;
  try {
    id = create.json('id');
  } catch (_) {
    id = null;
  }

  if (id) {
    const updatedSlug = `${slug}-updated`;
    updateCommunity(id, updatedSlug);
    deleteCommunity(id);
  }

}
