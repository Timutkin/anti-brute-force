import http from 'k6/http';
import { check } from 'k6';

// Generate 100 different users
let users = [];
for (let i = 1; i <= 100; i++) {
  users.push({
    login: `user${i}`,
    password: `pass${i}`,
    ip: `192.168.1.${i}`,
  });
}

export let options = {
  scenarios: {
    constant_load: {
      executor: 'constant-vus',
      vus: 1000,
      duration: '5m',
    },
  },
  thresholds: {
    checks: ['rate>0.99'], // At least 99% of all checks should pass
  },
};

export default function () {
  let user = users[__VU % 100];
  let payload = JSON.stringify({
    login: user.login,
    password: user.password,
    ip: user.ip,
  });
  let params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };
  let response = http.post('http://localhost:8081/allows', payload, params);
  check(response, {
    'status is 200 or 429': (r) => r.status === 200 || r.status === 429,
  });
}