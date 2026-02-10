import { beforeEach, describe, expect, it, vi, type Mock } from 'vitest';

type ClientModule = typeof import('./client');
type Middleware = {
  onRequest: (ctx: { request: Request }) => Promise<Request>;
  onResponse: (ctx: { request: Request; response: Response }) => Promise<Response>;
};

const dummyUser = {
  id: 'u-1',
  email: 'user@example.com',
  first_name: 'Test',
  last_name: 'User',
  role: 'employee' as const,
  is_active: true,
};

function makeFetchResponse(options?: {
  ok?: boolean;
  status?: number;
  jsonData?: unknown;
  blobData?: Blob;
}) {
  const {
    ok = true,
    status = 200,
    jsonData = { ok: true },
    blobData = new Blob(['ok'], { type: 'application/octet-stream' }),
  } = options || {};

  return {
    ok,
    status,
    json: vi.fn().mockResolvedValue(jsonData),
    blob: vi.fn().mockResolvedValue(blobData),
  };
}

async function loadClientWithMockedOpenAPIFetch(): Promise<{
  mod: ClientModule;
  useAuthStore: typeof import('@/stores/authStore').useAuthStore;
  middleware: Middleware;
}> {
  vi.resetModules();
  let capturedMiddleware: Middleware | undefined;

  vi.doMock('openapi-fetch', () => ({
    default: vi.fn(() => ({
      use: (mw: Middleware) => {
        capturedMiddleware = mw;
      },
    })),
  }));

  const mod = await import('./client');
  const { useAuthStore } = await import('@/stores/authStore');
  if (!capturedMiddleware) {
    throw new Error('middleware was not registered');
  }

  return { mod, useAuthStore, middleware: capturedMiddleware };
}

function collectApiFunctions(
  value: unknown,
  path: string[] = [],
): Array<{ path: string; fn: (...args: unknown[]) => Promise<unknown> }> {
  if (!value || typeof value !== 'object') return [];

  const entries = Object.entries(value as Record<string, unknown>);
  const out: Array<{ path: string; fn: (...args: unknown[]) => Promise<unknown> }> = [];

  for (const [key, child] of entries) {
    const nextPath = [...path, key];
    if (typeof child === 'function') {
      out.push({ path: nextPath.join('.'), fn: child as (...args: unknown[]) => Promise<unknown> });
      continue;
    }
    out.push(...collectApiFunctions(child, nextPath));
  }
  return out;
}

function buildArgs(path: string, length: number): unknown[] {
  if (path.endsWith('uploadReceipt')) {
    return [new File(['x'], 'receipt.txt', { type: 'text/plain' })];
  }
  if (path.endsWith('uploadDocument')) {
    const fd = new FormData();
    fd.append('file', new File(['x'], 'doc.txt', { type: 'text/plain' }));
    return [fd];
  }

  return Array.from({ length }, (_, idx) => {
    if (idx === 0) return 'id';
    if (idx === 1) return 'id2';
    return { sample: true };
  });
}

describe('client.ts branch coverage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    global.fetch = vi.fn() as unknown as typeof fetch;
    window.localStorage.clear();
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  it('covers apiClient middleware onRequest/onResponse branches', async () => {
    const { useAuthStore, middleware } = await loadClientWithMockedOpenAPIFetch();

    useAuthStore.getState().setAuth(dummyUser, 'access-token', 'refresh-token');

    const request = new Request('http://example.com/protected', {
      headers: new Headers({ 'X-Test': '1' }),
    });
    const requested = await middleware.onRequest({ request });
    expect(requested.headers.get('Authorization')).toBe('Bearer access-token');

    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({
        ok: true,
        status: 200,
        jsonData: {
          access_token: 'new-access',
          refresh_token: 'new-refresh',
          user: dummyUser,
        },
      }),
    );
    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: true, status: 200, jsonData: { retried: true } }),
    );

    const retried = await middleware.onResponse({
      request: new Request('http://example.com/protected', {
        headers: new Headers({ 'X-Test': '1' }),
      }),
      response: new Response(null, { status: 401 }),
    });
    expect(retried.status).toBe(200);

    useAuthStore.getState().logout();
    const noTokenReq = new Request('http://example.com/public');
    const noTokenOut = await middleware.onRequest({ request: noTokenReq });
    expect(noTokenOut.headers.get('Authorization')).toBeNull();

    const passthrough = await middleware.onResponse({
      request: new Request('http://example.com/public'),
      response: new Response(null, { status: 200 }),
    });
    expect(passthrough.status).toBe(200);

    const logoutSpy = vi.spyOn(useAuthStore.getState(), 'logout');
    (global.fetch as Mock).mockImplementation(() => {
      throw new Error('refresh call should not happen without refresh token');
    });

    const failed = await middleware.onResponse({
      request: new Request('http://example.com/protected'),
      response: new Response(null, { status: 401 }),
    });
    expect(failed.status).toBe(401);
    expect(logoutSpy).toHaveBeenCalled();
  });

  it('covers fetchWithAuth 401/refresh/retry branches', async () => {
    const { mod, useAuthStore } = await loadClientWithMockedOpenAPIFetch();
    const { api } = mod;

    useAuthStore.getState().setAuth(dummyUser, 'expired-token', 'refresh-token');

    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: false, status: 401, jsonData: { message: 'expired' } }),
    );
    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({
        ok: true,
        status: 200,
        jsonData: { access_token: 'new-token', refresh_token: 'r2', user: dummyUser },
      }),
    );
    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: true, status: 204, jsonData: null }),
    );

    const logoutResult = await api.auth.logout();
    expect(logoutResult).toBeNull();
    expect(useAuthStore.getState().accessToken).toBe('new-token');

    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: false, status: 401, jsonData: { message: 'expired' } }),
    );
    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({
        ok: true,
        status: 200,
        jsonData: { access_token: 'new-token-2', refresh_token: 'r3', user: dummyUser },
      }),
    );
    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: false, status: 400, jsonData: { message: 'retry failed' } }),
    );
    await expect(api.auth.logout()).rejects.toThrow('retry failed');

    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: false, status: 401, jsonData: { message: 'expired' } }),
    );
    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({
        ok: true,
        status: 200,
        jsonData: { access_token: 'new-token-3', refresh_token: 'r4', user: dummyUser },
      }),
    );
    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: false, status: 401, jsonData: { message: 'still unauthorized' } }),
    );
    await expect(api.auth.logout()).rejects.toThrow('Unauthorized');

    useAuthStore.getState().logout();
    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: false, status: 401, jsonData: { message: 'expired' } }),
    );
    await expect(api.auth.logout()).rejects.toThrow('Unauthorized');

    useAuthStore.getState().setAuth(dummyUser, 'expired-again', 'refresh-token');
    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: false, status: 401, jsonData: { message: 'expired' } }),
    );
    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: true, status: 200, jsonData: {} }),
    );
    await expect(api.auth.logout()).rejects.toThrow('Unauthorized');
  });

  it('covers refresh concurrency and catch branch', async () => {
    const { mod, useAuthStore } = await loadClientWithMockedOpenAPIFetch();
    const { api } = mod;

    useAuthStore.getState().setAuth(dummyUser, 'old-token', 'refresh-token');

    let refreshCalls = 0;
    let protected401Calls = 0;
    let retryCalls = 0;

    (global.fetch as Mock).mockImplementation(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes('/auth/refresh')) {
        refreshCalls += 1;
        await new Promise((resolve) => setTimeout(resolve, 10));
        return makeFetchResponse({
          ok: true,
          status: 200,
          jsonData: { access_token: 'concurrent-token', refresh_token: 'r5', user: dummyUser },
        });
      }
      if (url.includes('/auth/login') && protected401Calls < 2) {
        protected401Calls += 1;
        return makeFetchResponse({ ok: false, status: 401, jsonData: { message: 'expired' } });
      }
      retryCalls += 1;
      return makeFetchResponse({ ok: true, status: 200, jsonData: { ok: true } });
    });

    await Promise.all([
      api.auth.login({ email: 'a@example.com', password: 'x' }),
      api.auth.login({ email: 'b@example.com', password: 'x' }),
    ]);

    expect(refreshCalls).toBe(1);
    expect(retryCalls).toBe(2);

    useAuthStore.getState().setAuth(dummyUser, 'expired', 'refresh-token');
    (global.fetch as Mock).mockImplementation(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes('/auth/login')) {
        return makeFetchResponse({ ok: false, status: 401, jsonData: { message: 'expired' } });
      }
      throw new Error('network error');
    });

    await expect(api.auth.login({ email: 'c@example.com', password: 'x' })).rejects.toThrow('Unauthorized');
  });

  it('covers remaining short-circuit and upload branches', async () => {
    const { mod, useAuthStore } = await loadClientWithMockedOpenAPIFetch();
    const { api } = mod;

    // line 26: refresh HTTP not ok branch
    useAuthStore.getState().setAuth(dummyUser, 'expired', 'refresh-token');
    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: false, status: 401, jsonData: { message: 'expired' } }),
    );
    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: false, status: 500, jsonData: { message: 'refresh failed' } }),
    );
    await expect(api.auth.logout()).rejects.toThrow('Unauthorized');

    // lines 31/33: fallback side of `data.user || store.user` and `data.refresh_token || refreshToken`
    useAuthStore.getState().setAuth(dummyUser, 'expired2', 'keep-refresh');
    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: false, status: 401, jsonData: { message: 'expired' } }),
    );
    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: true, status: 200, jsonData: { access_token: 'only-access' } }),
    );
    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: true, status: 200, jsonData: { ok: true } }),
    );
    await api.auth.logout();
    expect(useAuthStore.getState().user).toEqual(dummyUser);
    expect(useAuthStore.getState().refreshToken).toBe('keep-refresh');

    // line 119: fallback error message branch after retry non-401 error
    useAuthStore.getState().setAuth(dummyUser, 'expired3', 'refresh-token');
    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: false, status: 401, jsonData: { message: 'expired' } }),
    );
    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: true, status: 200, jsonData: { access_token: 'new-token' } }),
    );
    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: false, status: 400, jsonData: {} }),
    );
    await expect(api.auth.logout()).rejects.toThrow(/API/);

    // line 143: fetchWithAuthBlob token false side
    useAuthStore.getState().logout();
    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: true, status: 200, blobData: new Blob(['x']) }),
    );
    await api.export.overtime({ start_date: '2025-01-01', end_date: '2025-01-31' });

    // lines 427/431: uploadReceipt token false side + error branch
    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: false, status: 400, jsonData: { message: 'bad upload' } }),
    );
    await expect(
      api.expenses.uploadReceipt(new File(['x'], 'receipt.txt', { type: 'text/plain' })),
    ).rejects.toThrow('Upload failed');

    // line 571: data || {} fallback side
    useAuthStore.getState().setAuth(dummyUser, 'token', 'refresh-token');
    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: true, status: 200, jsonData: { ok: true } }),
    );
    await api.hr.completeTraining('training-id');

    // lines 610/613: uploadDocument token false side + error branch
    useAuthStore.getState().logout();
    const fd = new FormData();
    fd.append('file', new File(['x'], 'doc.txt', { type: 'text/plain' }));
    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: false, status: 418, jsonData: { message: 'teapot' } }),
    );
    await expect(api.hr.uploadDocument(fd)).rejects.toThrow('HTTP 418');
  });

  it('covers blob helper success/error and executes all api wrapper functions once', async () => {
    const { mod, useAuthStore } = await loadClientWithMockedOpenAPIFetch();
    const { api } = mod;

    useAuthStore.getState().setAuth(dummyUser, 'token', 'refresh-token');

    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: true, status: 200, blobData: new Blob(['csv']) }),
    );
    await expect(
      api.export.attendance({ start_date: '2025-01-01', end_date: '2025-01-31' }),
    ).resolves.toBeInstanceOf(Blob);

    (global.fetch as Mock).mockResolvedValueOnce(
      makeFetchResponse({ ok: false, status: 500, jsonData: { message: 'blob error' } }),
    );
    await expect(
      api.export.projects({ start_date: '2025-01-01', end_date: '2025-01-31' }),
    ).rejects.toThrow();

    (global.fetch as Mock).mockImplementation(async () =>
      makeFetchResponse({
        ok: true,
        status: 200,
        jsonData: { ok: true },
        blobData: new Blob(['ok']),
      }),
    );

    const allFns = collectApiFunctions(api);
    for (const { path, fn } of allFns) {
      const args = buildArgs(path, fn.length);
      await expect(fn(...args)).resolves.toBeDefined();
    }

    expect(allFns.length).toBeGreaterThan(150);

    await api.attendance.clockOut();
    expect(global.fetch).toHaveBeenCalled();
  });
});
