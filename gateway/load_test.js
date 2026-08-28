import http from 'k6/http';
import { check, group } from 'k6';

export const options = {
  scenarios: {
    marks_test: {
      executor: 'constant-arrival-rate',
      rate: 50,
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 50,
      maxVUs: 1000,
      exec: 'runMarksCheck',
    },
    users_test: {
      executor: 'constant-arrival-rate',
      rate: 30,
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 40,
      maxVUs: 1000,
      exec: 'runUsersCheck',
    },
  },
};

const BASE_URL = 'http://localhost:8080';

function fastUUID() {
  return Math.random().toString(36).substring(2, 10) + 
         Math.random().toString(36).substring(2, 10);
}

function getParams(scenarioTag, extraHeaders = {}) {
  return {
    headers: Object.assign({ 'Content-Type': 'application/json' }, extraHeaders),
    tags: { scenario: scenarioTag },
  };
}

export function runUsersCheck() {
  group('Users Scenario', function () {
    const nickname = fastUUID();
    const password = 'easiset';
    const payload = JSON.stringify({ nickname, password });
    const uParams = getParams('users');

    // 1. reg
    const regRes = http.post(`${BASE_URL}/api/register`, payload, uParams);
    check(regRes, { 'users: reg status 200/201': (r) => r.status === 200 || r.status === 201 });

    // 2. log — successful
    const logRes = http.post(`${BASE_URL}/api/login`, payload, uParams);
    check(logRes, { 'users: log status 200': (r) => r.status === 200 });

    // 3. del
    const delParams = getParams('users', { 'X-Confirm-Password': password });
    const delRes = http.del(`${BASE_URL}/api/delete`, null, delParams);
    check(delRes, { 'users: del status 200/204': (r) => r.status === 200 || r.status === 204 });

    // 4. rep_log — error
    const repLogRes = http.post(`${BASE_URL}/api/login`, payload, uParams);
    check(repLogRes, { 'users: rep_log error (4xx/500)': (r) => r.status >= 400 });

    // 5. rep_del — error
    const repDelRes = http.del(`${BASE_URL}/api/delete`, null, delParams);
    check(repDelRes, { 'users: rep_del error (4xx/500)': (r) => r.status >= 400 });
  });
}

export function runMarksCheck() {
  group('Marks Scenario', function () {
    const nickname = fastUUID();
    const password = 'easiset';
    const mParams = getParams('marks');

    // 1. reg (нужен для создания юзера под метки)
    const regPayload = JSON.stringify({ nickname, password });
    const regRes = http.post(`${BASE_URL}/api/register`, regPayload, mParams);
    check(regRes, { 'marks: reg status 200/201': (r) => r.status === 200 || r.status === 201 });

    // 2. new
    const markNewPayload = JSON.stringify({
      latitude: 80,
      longitude: 80,
      Comment: 'default',
    });
    const markNewRes = http.post(`${BASE_URL}/api/marks/new`, markNewPayload, mParams);
    check(markNewRes, { 'marks: new status 200/201': (r) => r.status === 200 || r.status === 201 });

    // 3. get
    const markGetPayload = JSON.stringify({
      latitude: 80,
      longitude: 80,
    });
    const markGetRes = http.post(`${BASE_URL}/api/marks/get`, markGetPayload, mParams);
    check(markGetRes, { 'marks: get status 200': (r) => r.status === 200 });
  });
}
