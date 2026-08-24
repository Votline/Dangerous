import http from 'k6/http';
import { check } from 'k6';

export const options = {
  scenarios: {
    marks_test: {
      executor: 'constant-vus',
      vus: 50,
      duration: '30s',
      exec: 'runMarksCheck',
    },
    users_test: {
      executor: 'constant-vus',
      vus: 50,
      duration: '30s',
      exec: 'runUsersCheck',
    },
  },
};

const BASE_URL = 'http://localhost:8080';

function fastUUID() {
  return Math.random().toString(36).substring(2, 10) + 
         Math.random().toString(36).substring(2, 10);
}

const params = {
  headers: {
    'Content-Type': 'application/json',
  },
};

export function runUsersCheck() {
  const nickname = fastUUID();
  const password = 'easiset';
  const payload = JSON.stringify({ nickname, password });

  // 1. reg
  const regRes = http.post(`${BASE_URL}/api/register`, payload, params);
  check(regRes, { 'users: reg status 200/201': (r) => r.status === 200 || r.status === 201 });

  // 2. log — successfull
  const logRes = http.post(`${BASE_URL}/api/login`, payload, params);
  check(logRes, { 'users: log status 200': (r) => r.status === 200 });

  // 3. del
  const delParams = {
    headers: {
      'Content-Type': 'application/json',
      'X-Confirm-Password': password,
    },
  };
  const delRes = http.del(`${BASE_URL}/api/delete`, null, delParams);
  check(delRes, { 'users: del status 200/204': (r) => r.status === 200 || r.status === 204 });

  // 4. rep_log — error
  const repLogRes = http.post(`${BASE_URL}/api/login`, payload, params);
  check(repLogRes, { 'users: rep_log error (4xx/500)': (r) => r.status >= 400 });

  // 5. rep_del — error
  const repDelRes = http.del(`${BASE_URL}/api/delete`, null, delParams);
  check(repDelRes, { 'users: rep_del error (4xx/500)': (r) => r.status >= 400 });
}

export function runMarksCheck() {
  const nickname = fastUUID();
  const password = 'easiset';

  // 1. reg
  const regPayload = JSON.stringify({ nickname, password });
  const regRes = http.post(`${BASE_URL}/api/register`, regPayload, params);
  check(regRes, { 'marks: reg status 200/201': (r) => r.status === 200 || r.status === 201 });

  // 2. new
  const markNewPayload = JSON.stringify({
    latitude: 80,
    longitude: 80,
    Comment: 'default',
  });
  const markNewRes = http.post(`${BASE_URL}/api/marks/new`, markNewPayload, params);
  check(markNewRes, { 'marks: new status 200/201': (r) => r.status === 200 || r.status === 201 });

  // 3. get
  const markGetPayload = JSON.stringify({
    latitude: 80,
    longitude: 80,
  });
  const markGetRes = http.post(`${BASE_URL}/api/marks/get`, markGetPayload, params);
  check(markGetRes, { 'marks: get status 200': (r) => r.status === 200 });
}
