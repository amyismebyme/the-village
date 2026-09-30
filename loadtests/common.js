import http from 'k6/http';
import { check, fail } from 'k6';

export const BASE_URL = (__ENV.BASE_URL || 'http://localhost:8080').replace(/\/$/, '');
export const REQUEST_TIMEOUT = __ENV.K6_REQUEST_TIMEOUT || '5s';
export const COMMUNITY_COUNT = Number(__ENV.K6_COMMUNITY_COUNT || 20);

function requestOptions(endpoint) {
  return {
    timeout: REQUEST_TIMEOUT,
    tags: {
      endpoint,
    },
  };
}

export function getHealth() {
  return http.get(`${BASE_URL}/health`, requestOptions('health'));
}

export function getReady() {
  return http.get(`${BASE_URL}/ready`, requestOptions('ready'));
}

export function getCommunities() {
  return http.get(
    `${BASE_URL}/api/v1/communities?limit=20&offset=0`,
    requestOptions('community_list'),
  );
}

export function getCommunity(id) {
  return http.get(
    `${BASE_URL}/api/v1/communities/${id}`,
    requestOptions('community_get'),
  );
}

export function createCommunity(index, prefix = 'k6') {
  const unique = `${Date.now()}-${index}-${Math.floor(Math.random() * 1000000)}`;
  const payload = JSON.stringify({
    name: `${prefix} Community ${unique}`,
    slug: `${prefix}-community-${unique}`,
    description: 'Temporary k6 performance-test fixture.',
    external_source: 'k6',
  });

  return http.post(
    `${BASE_URL}/api/v1/communities`,
    payload,
    {
      timeout: REQUEST_TIMEOUT,
      headers: { 'Content-Type': 'application/json' },
      tags: { endpoint: 'community_create' },
    },
  );
}

export function setupCommunities() {
  const ids = [];

  for (let i = 0; i < COMMUNITY_COUNT; i += 1) {
    const response = createCommunity(i);

    const ok = check(response, {
      'fixture community created': (r) => r.status === 201,
    });

    if (!ok) {
      fail(`k6 fixture setup failed with status ${response.status}`);
    }

    const body = response.json();
    if (!body || !body.id) {
      fail('k6 fixture setup returned no community id');
    }

    ids.push(body.id);
  }

  return { ids };
}

export function teardownCommunities(data) {
  if (!data || !data.ids) {
    return;
  }

  for (const id of data.ids) {
    const response = http.del(
      `${BASE_URL}/api/v1/communities/${id}`,
      null,
      {
        timeout: REQUEST_TIMEOUT,
        tags: { endpoint: 'community_delete' },
      },
    );

    check(response, {
      'fixture community deleted': (r) => r.status === 204 || r.status === 404,
    });
  }
}
