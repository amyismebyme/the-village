import http from 'k6/http';
import { check } from 'k6';

import { DEFAULT_HEADERS, url } from './config.js';

export function getHealth() {
  const response = http.get(url('/health'), {
    tags: { route: '/health', operation: 'liveness' },
  });

  check(response, {
    'GET /health returns 200': (r) => r.status === 200,
  });

  return response;
}

export function getReady() {
  const response = http.get(url('/ready'), {
    tags: { route: '/ready', operation: 'readiness' },
  });

  check(response, {
    'GET /ready returns 200': (r) => r.status === 200,
  });

  return response;
}


export function getStatus() {
  const response = http.get(url('/status'), {
    tags: { route: '/status', operation: 'status' },
  });

  check(response, {
    'GET /status returns 200': (r) => r.status === 200,
  });

  return response;
}

export function listCommunities(limit = 20, offset = 0) {
  const response = http.get(
    url(`/api/v1/communities?limit=${limit}&offset=${offset}`),
    {
      tags: { route: '/api/v1/communities', operation: 'list' },
    },
  );

  check(response, {
    'GET /api/v1/communities returns 200': (r) => r.status === 200,
    'community list response is JSON': (r) =>
      r.headers['Content-Type']?.includes('application/json'),
  });

  return response;
}

export function getCommunity(id) {
  const response = http.get(url(`/api/v1/communities/${id}`), {
    tags: { route: '/api/v1/communities/{id}', operation: 'get' },
  });

  check(response, {
    'GET community returns 200': (r) => r.status === 200,
  });

  return response;
}

export function createCommunity(name, slug) {
  const response = http.post(
    url('/api/v1/communities'),
    JSON.stringify({
      name,
      slug,
      description: 'Temporary k6 performance smoke test resource',
      external_source: 'k6',
    }),
    {
      headers: DEFAULT_HEADERS,
      tags: { route: '/api/v1/communities', operation: 'create' },
    },
  );

  check(response, {
    'POST community returns 201': (r) => r.status === 201,
  });

  return response;
}

export function updateCommunity(id, slug) {
  const response = http.put(
    url(`/api/v1/communities/${id}`),
    JSON.stringify({
      name: 'K6 Performance Smoke Test Updated',
      slug,
      description: 'Updated temporary k6 performance smoke test resource',
      external_source: 'k6',
    }),
    {
      headers: DEFAULT_HEADERS,
      tags: { route: '/api/v1/communities/{id}', operation: 'update' },
    },
  );

  check(response, {
    'PUT community returns 200': (r) => r.status === 200,
  });

  return response;
}

export function deleteCommunity(id) {
  const response = http.del(url(`/api/v1/communities/${id}`), null, {
    tags: { route: '/api/v1/communities/{id}', operation: 'delete' },
  });

  check(response, {
    'DELETE community returns 204': (r) => r.status === 204,
  });

  return response;
}
