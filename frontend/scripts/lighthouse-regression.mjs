import process from 'node:process';
import { chromium } from '@playwright/test';
import lighthouse from 'lighthouse';
import * as chromeLauncher from 'chrome-launcher';
import { build, preview } from 'vite';

const baseURL = process.env.LIGHTHOUSE_BASE_URL ?? 'http://127.0.0.1:4173';
const paths = (process.env.LIGHTHOUSE_PATHS ?? '/login,/').split(',').map((p) => p.trim()).filter(Boolean);

const defaultThresholds = {
  performanceScore: Number(process.env.LH_MIN_PERFORMANCE_SCORE ?? 0.7),
  lcp: Number(process.env.LH_MAX_LCP_MS ?? 4000),
  cls: Number(process.env.LH_MAX_CLS ?? 0.1),
  inp: Number(process.env.LH_MAX_INP_MS ?? 500),
  tbt: Number(process.env.LH_MAX_TBT_MS ?? 600),
};

const thresholdsByPath = {
  '/login': {
    performanceScore: Number(process.env.LH_LOGIN_MIN_PERFORMANCE_SCORE ?? 0.55),
    lcp: Number(process.env.LH_LOGIN_MAX_LCP_MS ?? 9000),
    cls: Number(process.env.LH_LOGIN_MAX_CLS ?? 0.1),
    inp: Number(process.env.LH_LOGIN_MAX_INP_MS ?? 600),
    tbt: Number(process.env.LH_LOGIN_MAX_TBT_MS ?? 700),
  },
  '/': {
    performanceScore: Number(process.env.LH_HOME_MIN_PERFORMANCE_SCORE ?? 0.75),
    lcp: Number(process.env.LH_HOME_MAX_LCP_MS ?? 3000),
    cls: Number(process.env.LH_HOME_MAX_CLS ?? 0.1),
    inp: Number(process.env.LH_HOME_MAX_INP_MS ?? 700),
    tbt: Number(process.env.LH_HOME_MAX_TBT_MS ?? 1000),
  },
};

function resolveThresholds(url) {
  const pathname = new URL(url).pathname;
  return thresholdsByPath[pathname] ?? defaultThresholds;
}

/**
 * @param {number | undefined} value
 */
function formatMetric(value) {
  if (typeof value !== 'number' || Number.isNaN(value)) {
    return 'n/a';
  }
  return value.toFixed(2);
}

function sleep(ms) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}

function parseServerTarget(base) {
  const url = new URL(base);
  const port = Number(url.port || '3000');
  const host = url.hostname || '127.0.0.1';
  return { host, port };
}

async function waitForServerReady(base, timeoutMs = 60000) {
  const startedAt = Date.now();
  while (Date.now() - startedAt < timeoutMs) {
    try {
      const res = await fetch(base);
      if (res.ok || res.status < 500) {
        return;
      }
    } catch {
      // retry
    }
    await sleep(500);
  }
  throw new Error(`frontend server did not become ready within ${timeoutMs}ms: ${base}`);
}

async function stopFrontendServer(child) {
  if (!child) {
    return;
  }
  await child.close();
}

/**
 * @param {import('lighthouse').RunnerResult['lhr']} lhr
 * @param {string} url
 */
function validateMetrics(lhr, url) {
  const thresholds = resolveThresholds(url);
  const perfScore = lhr.categories.performance?.score ?? 0;
  const lcp = lhr.audits['largest-contentful-paint']?.numericValue;
  const cls = lhr.audits['cumulative-layout-shift']?.numericValue;
  const inp =
    lhr.audits['interaction-to-next-paint']?.numericValue ??
    lhr.audits['max-potential-fid']?.numericValue;
  const tbt = lhr.audits['total-blocking-time']?.numericValue;

  const violations = [];
  if (perfScore < thresholds.performanceScore) {
    violations.push(
      `${url}: performance score ${formatMetric(perfScore)} < ${thresholds.performanceScore.toFixed(2)}`,
    );
  }
  if (typeof lcp === 'number' && lcp > thresholds.lcp) {
    violations.push(`${url}: LCP ${formatMetric(lcp)}ms > ${thresholds.lcp}ms`);
  }
  if (typeof cls === 'number' && cls > thresholds.cls) {
    violations.push(`${url}: CLS ${formatMetric(cls)} > ${thresholds.cls}`);
  }
  if (typeof inp === 'number' && inp > thresholds.inp) {
    violations.push(`${url}: INP ${formatMetric(inp)}ms > ${thresholds.inp}ms`);
  }
  if (typeof tbt === 'number' && tbt > thresholds.tbt) {
    violations.push(`${url}: TBT ${formatMetric(tbt)}ms > ${thresholds.tbt}ms`);
  }

  return {
    url,
    performanceScore: perfScore,
    lcp,
    cls,
    inp,
    tbt,
    violations,
  };
}

async function main() {
  const shouldStartServer = process.env.LIGHTHOUSE_SKIP_SERVER_START !== '1';
  const target = parseServerTarget(baseURL);
  let frontendServer = null;
  try {
    if (shouldStartServer) {
      await build({ logLevel: 'error' });
      frontendServer = await preview({
        logLevel: 'error',
        preview: {
          host: target.host,
          port: target.port,
          strictPort: true,
        },
      });
      await waitForServerReady(baseURL);
    }

    const chromePath = chromium.executablePath();
    const chrome = await chromeLauncher.launch({
      chromePath,
      logLevel: 'error',
      chromeFlags: ['--headless=new', '--no-sandbox', '--disable-dev-shm-usage'],
    });

    /** @type {Array<ReturnType<typeof validateMetrics>>} */
    const results = [];
    try {
      for (const path of paths) {
        const url = new URL(path, baseURL).toString();
        const run = await lighthouse(
          url,
          {
            port: chrome.port,
            logLevel: 'error',
            output: 'json',
            onlyCategories: ['performance'],
            disableStorageReset: true,
          },
          undefined,
        );

        if (!run?.lhr) {
          throw new Error(`failed to run lighthouse for ${url}`);
        }
        results.push(validateMetrics(run.lhr, url));
      }
    } finally {
      try {
        await chrome.kill();
      } catch (error) {
        if (!error || typeof error !== 'object' || !('code' in error) || error.code !== 'EPERM') {
          throw error;
        }
      }
    }

    console.table(
      results.map((row) => ({
        url: row.url,
        performanceScore: formatMetric(row.performanceScore),
        lcpMs: formatMetric(row.lcp),
        cls: formatMetric(row.cls),
        inpMs: formatMetric(row.inp),
        tbtMs: formatMetric(row.tbt),
      })),
    );

    const violations = results.flatMap((row) => row.violations);
    if (violations.length > 0) {
      console.error('Lighthouse regression detected:');
      for (const violation of violations) {
        console.error(`- ${violation}`);
      }
      process.exit(1);
    }
  } finally {
    await stopFrontendServer(frontendServer);
  }
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
