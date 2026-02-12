export function getBaseURL() {
  return __ENV.K6_BASE_URL || "http://localhost:8080";
}

export function getEnvInt(name, defaultValue) {
  const raw = __ENV[name];
  if (!raw) {
    return defaultValue;
  }

  const parsed = Number.parseInt(raw, 10);
  if (Number.isNaN(parsed)) {
    return defaultValue;
  }

  return parsed;
}

export function getEnvDuration(name, defaultValue) {
  return __ENV[name] || defaultValue;
}

export function nowMillis() {
  return Date.now();
}
