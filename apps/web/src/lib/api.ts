import type { DashboardPayload, GanttPayload, Order, SessionPayload } from "./types";

const BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8014";

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

export function apiUrl(path: string) {
  return `${BASE}${path}`;
}

type Query = Record<string, string | undefined>;

function qs(q: Query) {
  const p = new URLSearchParams();
  for (const [k, v] of Object.entries(q)) {
    if (v) p.set(k, v);
  }
  const s = p.toString();
  return s ? `?${s}` : "";
}

async function req<T>(path: string, user: string, init?: RequestInit): Promise<T> {
  let res: Response;
  try {
    res = await fetch(apiUrl(path), {
      ...init,
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        "X-Demo-User": user,
        ...(init?.headers ?? {}),
      },
      cache: "no-store",
    });
  } catch {
    throw new ApiError(0, "Loom API is unreachable. Start the Fiber server on port 8014.");
  }
  if (!res.ok) {
    let msg = `Request failed (${res.status})`;
    try {
      const body = (await res.json()) as { error?: string };
      if (body.error) msg = body.error;
    } catch {
      /* ignore */
    }
    throw new ApiError(res.status, msg);
  }
  return (await res.json()) as T;
}

export const api = {
  health: (user: string) => req<{ status: string; db: string; seeded: boolean; sim: string }>(`/health`, user),
  session: (user: string) => req<SessionPayload>(`/api/session${qs({ user })}`, user),
  gantt: (user: string, plant?: string) =>
    req<GanttPayload>(`/api/gantt${qs({ user, plant })}`, user),
  dashboard: (user: string, plant?: string) =>
    req<DashboardPayload>(`/api/dashboard${qs({ user, plant })}`, user),
  orders: (user: string, plant?: string) =>
    req<{ orders: Order[]; other_orders?: Order[]; focus?: string }>(`/api/orders${qs({ user, plant })}`, user),
  place: (
    user: string,
    body: {
      order_id: string;
      line_id: string;
      starts_at: string;
      ends_at: string;
      planned_qty: number;
      sop_state?: string;
    },
  ) => req<{ block: unknown }>(`/api/blocks`, user, { method: "POST", body: JSON.stringify(body) }),
  move: (
    user: string,
    id: string,
    body: { line_id?: string; starts_at: string; ends_at: string },
  ) => req<{ block: unknown }>(`/api/blocks/${id}`, user, { method: "PATCH", body: JSON.stringify(body) }),
  advance: (user: string, id: string) =>
    req<{ sop_state: string }>(`/api/blocks/${id}/sop`, user, {
      method: "POST",
      body: JSON.stringify({ action: "advance" }),
    }),
};
