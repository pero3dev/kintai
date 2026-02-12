import { sleep } from "k6";
import { getBaseURL, getEnvInt, nowMillis } from "../lib/config.js";
import { issueSingleAccessToken } from "../lib/auth.js";
import { callUsersMe } from "../lib/users_me.js";

export const options = {
  stages: [
    { duration: __ENV.K6_STAGE_NORMAL_DURATION || "30s", target: getEnvInt("K6_STAGE_NORMAL_VUS", 5) },
    { duration: __ENV.K6_STAGE_PEAK_DURATION || "30s", target: getEnvInt("K6_STAGE_PEAK_VUS", 20) },
    { duration: __ENV.K6_STAGE_SPIKE_DURATION || "20s", target: getEnvInt("K6_STAGE_SPIKE_VUS", 50) },
    { duration: __ENV.K6_STAGE_COOLDOWN_DURATION || "10s", target: 0 },
  ],
  thresholds: {
    http_req_failed: ["rate<0.02"],
    "http_req_duration{endpoint:users_me}": ["p(95)<350"],
  },
};

export function setup() {
  const baseURL = getBaseURL();
  const password = __ENV.K6_PASSWORD || "password123";
  const email = `${__ENV.K6_EMAIL_PREFIX || "k6-load-profile"}-${nowMillis()}@example.com`;
  const token = issueSingleAccessToken(baseURL, email, password);

  return {
    baseURL: baseURL,
    token: token,
  };
}

export default function (data) {
  callUsersMe(data.baseURL, data.token);
  sleep(getEnvInt("K6_THINK_TIME_MS", 50) / 1000);
}
