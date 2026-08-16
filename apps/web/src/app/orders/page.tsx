"use client";

import { AppShell } from "@/components/AppShell";
import { ErrorLoom, LoadingLoom } from "@/components/StatusStates";
import { api, ApiError } from "@/lib/api";
import { fmtQty } from "@/lib/format";
import type { Order, Plant, SessionPayload } from "@/lib/types";
import { useSearchParams } from "next/navigation";
import { Suspense, useCallback, useEffect, useState } from "react";

function OrdersBoard() {
  const params = useSearchParams();
  const user = params.get("user") || "kavya";
  const plant = (params.get("plant") || "TK").toUpperCase();
  const [session, setSession] = useState<SessionPayload | null>(null);
  const [orders, setOrders] = useState<Order[]>([]);
  const [others, setOthers] = useState<Order[]>([]);
  const [plants, setPlants] = useState<Plant[]>([]);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setError(null);
    try {
      const [s, o, g] = await Promise.all([
        api.session(user),
        api.orders(user, plant),
        api.gantt(user, plant),
      ]);
      setSession(s);
      setOrders(o.orders);
      setOthers(o.other_orders ?? []);
      setPlants(g.plants);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Orders failed to load.");
    }
  }, [user, plant]);

  useEffect(() => {
    void load();
  }, [load]);

  if (error && !session) {
    return <ErrorLoom message={error} onRetry={() => void load()} />;
  }
  if (!session) {
    return <LoadingLoom />;
  }

  return (
    <AppShell users={session.users} plants={plants} user={user} plant={plant}>
      {error ? <ErrorLoom message={error} onRetry={() => void load()} /> : null}
      <p className="muted">
        Place leftover quantity from the warp. Buyer, due date, and open pcs — not an ERP grid.
      </p>
      <div className="order-grid">
        {orders.map((o) => (
          <article className="order-card" key={o.id}>
            <div className="muted">{o.po_code}</div>
            <h3>
              {o.buyer} · {o.style_name}
            </h3>
            <div className="muted">
              {o.colorway} · {o.plant_code} · due {o.due_date}
            </div>
            <p>
              <strong>{fmtQty(o.qty)}</strong> pcs · {fmtQty(o.unplaced_qty)} still open
            </p>
            <div className="muted">{o.merch_owner}</div>
          </article>
        ))}
      </div>
      {others.length ? (
        <>
          <h2
            className="mt-4"
            style={{ fontFamily: "var(--font-condensed), sans-serif", letterSpacing: "0.12em" }}
          >
            Other houses
          </h2>
          <div className="order-grid">
            {others.map((o) => (
              <article className="order-card" key={o.id} style={{ opacity: 0.55 }}>
                <div className="muted">{o.po_code}</div>
                <h3>
                  {o.buyer} · {o.style_name}
                </h3>
                <div className="muted">{o.merch_owner}</div>
              </article>
            ))}
          </div>
        </>
      ) : null}
    </AppShell>
  );
}

export default function OrdersPage() {
  return (
    <Suspense fallback={<LoadingLoom />}>
      <OrdersBoard />
    </Suspense>
  );
}
