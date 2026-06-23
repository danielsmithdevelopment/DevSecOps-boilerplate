import http from 'k6/http';
import { sleep, check } from 'k6';

export const options = {
  vus: 10,
  duration: '1m',
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<500'],
  },
};

const TARGET = __ENV.K6_TARGET_URL || 'http://task-runner-1:8080';

export default function () {
  const res = http.get(`${TARGET}/health`);
  check(res, {
    'status is 200': (r) => r.status === 200,
  });
  sleep(1);
}
