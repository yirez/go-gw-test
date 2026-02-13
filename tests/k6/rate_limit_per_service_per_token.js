import http from "k6/http";
import { check, fail, sleep } from "k6";

const API_GW_BASE_URL = __ENV.API_GW_BASE_URL || "http://localhost:8085";
const AUTH_GW_BASE_URL = __ENV.AUTH_GW_BASE_URL || "http://localhost:8084";
const USERNAME = __ENV.AUTH_USERNAME || "user_all";
const PASSWORD = __ENV.AUTH_PASSWORD || "123";
const BURST_REQUESTS = Number(__ENV.BURST_REQUESTS || 8);
const USERS_PATH = __ENV.USERS_PATH || "/api/v1/users";
const ORDERS_PATH = __ENV.ORDERS_PATH || "/api/v1/orders";

export const options = {
  vus: 1,
  iterations: 1,
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

function alignToSecondStart() {
  const nowMs = Date.now();
  const msUntilNextSecond = 1000 - (nowMs % 1000);
  // Start shortly after the second flips, to keep burst in one rate window.
  sleep((msUntilNextSecond + 30) / 1000);
}

function burst(baseURL, path, token, count) {
  let okCount = 0;
  let tooManyCount = 0;
  let otherCount = 0;

  for (let i = 0; i < count; i += 1) {
    const response = http.get(`${baseURL}${path}`, {
      headers: authHeaders(token),
    });
    if (response.status === 200) {
      okCount += 1;
    } else if (response.status === 429) {
      tooManyCount += 1;
    } else {
      otherCount += 1;
    }
  }

  return {
    ok: okCount,
    tooMany: tooManyCount,
    other: otherCount,
  };
}

export default function () {
  const tokenA = login();
  const tokenB = login();

  // Test A: same token, different services in same second.
  alignToSecondStart();
  const userBurst = burst(API_GW_BASE_URL, USERS_PATH, tokenA, BURST_REQUESTS);
  const ordersAfterUserBurst = http.get(`${API_GW_BASE_URL}${ORDERS_PATH}`, {
    headers: authHeaders(tokenA),
  });

  // Test B: same service, different tokens in same second.
  alignToSecondStart();
  const tokenABurst = burst(API_GW_BASE_URL, USERS_PATH, tokenA, BURST_REQUESTS);
  const tokenBProbe = http.get(`${API_GW_BASE_URL}${USERS_PATH}`, {
    headers: authHeaders(tokenB),
  });

  const pass = check(
    {
      userBurst: userBurst,
      ordersAfterUserBurst: ordersAfterUserBurst,
      tokenABurst: tokenABurst,
      tokenBProbe: tokenBProbe,
    },
    {
      "A1: users burst has at least one 429": (r) => r.userBurst.tooMany > 0,
      "A2: users burst still has successful requests": (r) => r.userBurst.ok > 0,
      "A3: orders call is not rate limited by users burst": (r) => r.ordersAfterUserBurst.status !== 429,
      "B1: token A burst has at least one 429": (r) => r.tokenABurst.tooMany > 0,
      "B2: token B still allowed in same window": (r) => r.tokenBProbe.status === 200,
    },
  );

  console.log(
    JSON.stringify(
      {
        users_burst: userBurst,
        orders_after_users_burst_status: ordersAfterUserBurst.status,
        token_a_users_burst: tokenABurst,
        token_b_users_probe_status: tokenBProbe.status,
      },
      null,
      2,
    ),
  );

  if (!pass) {
    fail("rate-limit assertions failed");
  }
}
