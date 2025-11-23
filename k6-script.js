import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export const options = {
  stages: [
    { duration: '10s', target: 10 },
    { duration: '30s', target: 50 },
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<300'],
    http_req_failed: ['rate<0.01'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://host.docker.internal:8080';

export default function () {
  const teamName = `load-team-${randomString(5)}`;
  const authorID = `user-${randomString(5)}`;
  const prID = `pr-${randomString(8)}`;

  const teamPayload = JSON.stringify({
    team_name: teamName,
    members: [
      { user_id: authorID, username: "Author", is_active: true },
      { user_id: "rev1", username: "Reviewer1", is_active: true },
      { user_id: "rev2", username: "Reviewer2", is_active: true }
    ]
  });

  const params = { headers: { 'Content-Type': 'application/json' } };
  
  let res = http.post(`${BASE_URL}/team/add`, teamPayload, params);
  check(res, { 'team created': (r) => r.status === 201 || r.status === 400 });

  const prPayload = JSON.stringify({
    pull_request_id: prID,
    pull_request_name: "Load Test PR",
    author_id: authorID
  });

  res = http.post(`${BASE_URL}/pullRequest/create`, prPayload, params);
  
  check(res, { 
    'pr created': (r) => r.status === 201,
  });

  res = http.get(`${BASE_URL}/users/getReview?user_id=rev1`);
  check(res, { 'get reviews': (r) => r.status === 200 });

  sleep(1);
}