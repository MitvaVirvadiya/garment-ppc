"use client";

import { AppShell } from "@/components/AppShell";
import { BlockInspector } from "@/components/BlockInspector";
import { KpiStrip } from "@/components/KpiStrip";
import { LiveBoard } from "@/components/LiveBoard";
import { PlacePanel } from "@/components/PlacePanel";
import { EmptyLoom, ErrorLoom, LoadingLoom } from "@/components/StatusStates";
import { SnagRail } from "@/components/SnagRail";
import { WarpGantt } from "@/components/WarpGantt";
import { api, ApiError } from "@/lib/api";
import type { Block, Conflict, DashboardPayload, GanttPayload, Order, SessionPayload } from "@/lib/types";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useCallback, useEffect, useState } from "react";

function ControlRoom() {
  const params = useSearchParams();
  const router = useRouter();
  const user = params.get("user") || "kavya";
  const plant = (params.get("plant") || "TK").toUpperCase();

  const [session, setSession] = useState<SessionPayload | null>(null);
  const [gantt, setGantt] = useState<GanttPayload | null>(null);
  const [dash, setDash] = useState<DashboardPayload | null>(null);
  const [orders, setOrders] = useState<Order[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [selected, setSelected] = useState<Block | null>(null);
  const [snag, setSnag] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const refresh = useCallback(async () => {
    const [g, d, o] = await Promise.all([
      api.gantt(user, plant),
      api.dashboard(user, plant),
      api.orders(user, plant),
    ]);
    setGantt(g);
    setDash(d);
    setOrders(o.orders);
    setSelected((cur) => (cur ? (g.blocks.find((b) => b.id === cur.id) ?? cur) : null));
  }, [user, plant]);

  const load = useCallback(async () => {
    setError(null);
    try {
      const s = await api.session(user);
      setSession(s);
      const locked = s.session.locked_plant_id;
      if (locked) {
        const all = await api.gantt(user);
        const lockedPlant = all.plants.find((p) => p.id === locked);
        if (lockedPlant && lockedPlant.code !== plant) {
          router.replace(`/?user=${user}&plant=${lockedPlant.code}`);
          return;
        }
      }
      await refresh();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Could not load the warp.");
    }
  }, [user, plant, router, refresh]);

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    const id = window.setInterval(() => {
      void refresh().catch(() => undefined);
    }, 4000);
    return () => window.clearInterval(id);
  }, [refresh]);

  const tz = gantt?.plants.find((p) => p.code === plant)?.timezone ?? "Asia/Kolkata";

  async function move(id: string, starts: string, ends: string, lineId: string) {
    setBusy(true);
    try {
      await api.move(user, id, { line_id: lineId, starts_at: starts, ends_at: ends });
      await load();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Move failed.");
    } finally {
      setBusy(false);
    }
  }

  async function advance() {
    if (!selected) return;
    setBusy(true);
    try {
      await api.advance(user, selected.id);
      await load();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "SOP did not advance.");
    } finally {
      setBusy(false);
    }
  }

  const plants = gantt?.plants ?? [];
  const users = session?.users ?? [];

  return (
    <AppShell users={users} plants={plants} user={user} plant={plant}>
      {error ? <ErrorLoom message={error} onRetry={() => void load()} /> : null}
      {!gantt || !dash ? (
        <LoadingLoom />
      ) : (
        <>
          <KpiStrip dash={dash} session={gantt.session} />
          <SnagRail
            conflicts={gantt.conflicts}
            activeId={snag}
            tz={tz}
            onFocus={(c: Conflict) => {
              setSnag(c.id);
              const b = gantt.blocks.find((x) => c.block_ids.includes(x.id));
              if (b) setSelected(b);
            }}
          />
          {gantt.lines.length === 0 ? (
            <EmptyLoom hint="Run go run ./cmd/seed if this plant has no lines." />
          ) : (
            <WarpGantt
              from={gantt.from}
              lines={gantt.lines}
              blocks={gantt.blocks}
              conflicts={gantt.conflicts}
              session={gantt.session}
              tz={tz}
              selectedId={selected?.id ?? null}
              onSelect={setSelected}
              onMoved={move}
              highlightConflict={snag}
            />
          )}
          <div className="dock">
            <LiveBoard blocks={gantt.blocks} />
            <BlockInspector block={selected} session={gantt.session} tz={tz} onAdvance={advance} busy={busy} />
          </div>
          <PlacePanel
            session={gantt.session}
            orders={orders}
            lines={gantt.lines}
            busy={busy}
            onPlace={async (body) => {
              setBusy(true);
              try {
                await api.place(user, body);
                await load();
              } catch (e) {
                setError(e instanceof ApiError ? e.message : "Place failed.");
              } finally {
                setBusy(false);
              }
            }}
          />
        </>
      )}
    </AppShell>
  );
}

export default function Page() {
  return (
    <Suspense fallback={<LoadingLoom />}>
      <ControlRoom />
    </Suspense>
  );
}
