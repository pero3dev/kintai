import { getBaseURL, getEnvDuration, getEnvInt, nowMillis } from "../lib/config.js";
import { issueAccessTokens } from "../lib/auth.js";
import { callUsersMe } from "../lib/users_me.js";

const vus = getEnvInt("K6_VUS", 80);
const duration = getEnvDuration("K6_DURATION", "30s");

export const options = {
  vus: vus,
  duration: duration,
  thresholds: {
    http_req_failed: ["rate<0.01"],
    "http_req_duration{endpoint:users_me}": ["p(95)<500"],
  },
};

export function setup() {
  const baseURL = getBaseURL();
  const password = __ENV.K6_PASSWORD || "password123";
  const emailPrefix = __ENV.K6_EMAIL_PREFIX || "k6-high-concurrency";
  const tokenCount = getEnvInt("K6_TOKEN_USERS", vus);
  const tokens = issueAccessTokens(baseURL, tokenCount, emailPrefix, password, nowMillis());

  return {
    baseURL: baseURL,
    tokens: tokens,
  };
}

export default function (data) {
  const tokenIndex = (__VU - 1) % data.tokens.length;
  callUsersMe(data.baseURL, data.tokens[tokenIndex]);
}
