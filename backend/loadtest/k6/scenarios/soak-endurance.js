import { sleep } from "k6";
import { getBaseURL, getEnvDuration, getEnvInt, nowMillis } from "../lib/config.js";
import { issueSingleAccessToken } from "../lib/auth.js";
import { callUsersMe } from "../lib/users_me.js";

const vus = getEnvInt("K6_VUS", 12);
const duration = getEnvDuration("K6_DURATION", "10m");

export const options = {
  vus: vus,
  duration: duration,
  thresholds: {
    http_req_failed: ["rate<0.01"],
    "http_req_duration{endpoint:users_me}": ["p(95)<350"],
  },
};

export function setup() {
  const baseURL = getBaseURL();
  const password = __ENV.K6_PASSWORD || "password123";
  const email = `${__ENV.K6_EMAIL_PREFIX || "k6-soak"}-${nowMillis()}@example.com`;
  const token = issueSingleAccessToken(baseURL, email, password);

  return {
    baseURL: baseURL,
    token: token,
  };
}

export default function (data) {
  callUsersMe(data.baseURL, data.token);
  sleep(getEnvInt("K6_THINK_TIME_MS", 20) / 1000);
}
