import http from "k6/http";
import { check, fail } from "k6";

function jsonHeaders(extraHeaders) {
  return Object.assign({ "Content-Type": "application/json" }, extraHeaders || {});
}

function registerUser(baseURL, email, password) {
  const registerPayload = JSON.stringify({
    email: email,
    password: password,
    first_name: "K6",
    last_name: "Load",
  });

  const registerRes = http.post(`${baseURL}/api/v1/auth/register`, registerPayload, {
    headers: jsonHeaders(),
    tags: { endpoint: "auth_register" },
  });

  const ok = check(registerRes, {
    "register status is 201": (r) => r.status === 201,
  });

  if (!ok) {
    fail(`register failed status=${registerRes.status} body=${registerRes.body}`);
  }
}

function loginUser(baseURL, email, password) {
  const loginPayload = JSON.stringify({
    email: email,
    password: password,
  });

  const loginRes = http.post(`${baseURL}/api/v1/auth/login`, loginPayload, {
    headers: jsonHeaders(),
    tags: { endpoint: "auth_login" },
  });

  const ok = check(loginRes, {
    "login status is 200": (r) => r.status === 200,
    "login has access token": (r) => {
      const body = r.json();
      return body && body.access_token;
    },
  });

  if (!ok) {
    fail(`login failed status=${loginRes.status} body=${loginRes.body}`);
  }

  return loginRes.json("access_token");
}

export function issueSingleAccessToken(baseURL, email, password) {
  registerUser(baseURL, email, password);
  return loginUser(baseURL, email, password);
}

export function issueAccessTokens(baseURL, users, emailPrefix, password, timestampMillis) {
  const tokens = [];

  for (let i = 0; i < users; i += 1) {
    const email = `${emailPrefix}-${timestampMillis}-${i}@example.com`;
    registerUser(baseURL, email, password);
    tokens.push(loginUser(baseURL, email, password));
  }

  return tokens;
}
