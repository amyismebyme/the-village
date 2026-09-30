export const BASE_URL = (__ENV.K6_BASE_URL || 'http://host.docker.internal:8080').replace(/\/$/, '');

export function url(path) {
  return `${BASE_URL}${path}`;
}

export function uniqueSlug(prefix = 'k6') {
  return `${prefix}-${Date.now()}-${__VU}-${__ITER}`;
}

export const DEFAULT_HEADERS = {
  'Content-Type': 'application/json',
};
