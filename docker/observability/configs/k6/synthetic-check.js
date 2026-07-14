import http from 'k6/http';
import { check, sleep } from 'k6';

/**
 * Production-style synthetic monitor — low rate, critical-path checks.
 * Run continuously (e.g. k6 operator / cron) rather than as a load storm.
 */
export const options = {
  vus: 1,
  duration: '2m',
  thresholds: {
    checks: ['rate>0.99'],
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<1000'],
  },
};

const TARGET = __ENV.K6_TARGET_URL || 'http://task-runner-1:8080';

export default function () {
  const health = http.get(`${TARGET}/health`);
  check(health, {
    'health 200': (r) => r.status === 200,
  });

  // Extend with login / purchase journeys for your app.
  sleep(30);
}
