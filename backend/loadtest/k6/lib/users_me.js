import http from "k6/http";
import { check } from "k6";

export function callUsersMe(baseURL, accessToken) {
  const res = http.get(`${baseURL}/api/v1/users/me`, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
    tags: { endpoint: "users_me" },
  });

  check(res, {
    "users/me status is 200": (r) => r.status === 200,
  });

  return res;
}
