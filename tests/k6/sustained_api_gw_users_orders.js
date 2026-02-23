import http from "k6/http";
import { check, fail } from "k6";

const API_GW_BASE_URL = __ENV.API_GW_BASE_URL || "http://localhost:8085";
const AUTH_GW_BASE_URL = __ENV.AUTH_GW_BASE_URL || "http://localhost:8084";
const USERNAME = __ENV.AUTH_USERNAME || "user_all";
const PASSWORD = __ENV.AUTH_PASSWORD || "123";
const USERS_PATH = __ENV.USERS_PATH || "/api/v1/users";
const ORDERS_PATH = __ENV.ORDERS_PATH || "/api/v1/orders";

export const options = {
  scenarios: {
    users_sustained: {
      executor: "constant-arrival-rate",
      exec: "hitUsers",
      rate: 3,
      timeUnit: "1s",
      duration: "10m",
      preAllocatedVUs: 5,
      maxVUs: 20,
    },
    orders_sustained: {
      executor: "constant-arrival-rate",
      exec: "hitOrders",
      rate: 2,
      timeUnit: "1s",
      duration: "10m",
      preAllocatedVUs: 5,
      maxVUs: 20,
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.05"],
    http_req_duration: ["p(95)<1000"],
  },
};

function login() {
  const payload = JSON.stringify({ username: USERNAME, password: PASSWORD });
  const headers = { "Content-Type": "application/json" };
  const response = http.post(
    `${AUTH_GW_BASE_URL}/auth/login`,
    payload,
    { headers: headers },
  );

  const ok = check(response, {
    "login returns 200": (r) => r.status === 200,
    "login returns token": (r) => !!r.json("token"),
  });
  if (!ok) {
    fail(`login failed: status=${response.status} body=${response.body}`);
  }

  return response.json("token");
}

function authHeaders(token) {
  return {
    Authorization: `Bearer ${token}`,
  };
}

export function setup() {
  return { token: login() };
}

export function hitUsers(data) {
  const response = http.get(`${API_GW_BASE_URL}${USERS_PATH}`, {
    headers: authHeaders(data.token),
  });

  check(response, {
    "users status is 200": (r) => r.status === 200,
  });
}

export function hitOrders(data) {
  const response = http.get(`${API_GW_BASE_URL}${ORDERS_PATH}`, {
    headers: authHeaders(data.token),
  });

  check(response, {
    "orders status is 200": (r) => r.status === 200,
  });
}
