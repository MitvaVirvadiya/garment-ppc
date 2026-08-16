"use client";

import { addHours, fmtDay, fmtPct } from "@/lib/format";
import type { Block, Conflict, Line, Session } from "@/lib/types";
import { useMemo, useRef, useState } from "react";

const DAYS = 14;
const DAY_MS = 86_400_000;

function startOfLocalDay(d: Date) {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate());
}

export function WarpGantt({
  from,
  lines,
  blocks,
  conflicts,
  session,
  tz,
  selectedId,
  onSelect,
  onMoved,
  highlightConflict,
}: {
  from: string;
  lines: Line[];
  blocks: Block[];
  conflicts: Conflict[];
  session: Session;
  tz: string;
  selectedId: string | null;
  onSelect: (b: Block) => void;
  onMoved: (id: string, starts: string, ends: string, lineId: string) => Promise<void>;
  highlightConflict: string | null;
}) {
  const windowStart = useMemo(() => startOfLocalDay(new Date(from)), [from]);
  const days = useMemo(
    () => Array.from({ length: DAYS }, (_, i) => new Date(windowStart.getTime() + i * DAY_MS)),
    [windowStart],
  );
  const today = startOfLocalDay(new Date());
  const nowX = ((Date.now() - windowStart.getTime()) / DAY_MS) * 100;

  const [drag, setDrag] = useState<{
    id: string;
    originX: number;
    dxDays: number;
    originStart: string;
    originEnd: string;
  } | null>(null);
  const trackRefs = useRef<Record<string, HTMLDivElement | null>>({});

  const blocksByLine = useMemo(() => {
    const m = new Map<string, Block[]>();
    for (const b of blocks) {
      const arr = m.get(b.line_id) ?? [];
      arr.push(b);
      m.set(b.line_id, arr);
    }
    return m;
  }, [blocks]);

  const focusBlocks = new Set(
    highlightConflict ? (conflicts.find((c) => c.id === highlightConflict)?.block_ids ?? []) : [],
  );

  function xPct(iso: string) {
    return ((new Date(iso).getTime() - windowStart.getTime()) / (DAYS * DAY_MS)) * 100;
  }

  function widthPct(a: string, b: string) {
    return ((new Date(b).getTime() - new Date(a).getTime()) / (DAYS * DAY_MS)) * 100;
  }

  function dim(b: Block) {
    if (session.kpi_set === "merch" && session.focus_buyer && b.merch_owner !== session.focus_buyer) {
      return true;
    }
    if (focusBlocks.size && !focusBlocks.has(b.id)) return true;
    return false;
  }

  async function finishDrag() {
    if (!drag) return;
    const hours = drag.dxDays * 24;
    const nextStart = addHours(drag.originStart, hours);
    const nextEnd = addHours(drag.originEnd, hours);
    const id = drag.id;
    const lineId = blocks.find((b) => b.id === id)?.line_id;
    setDrag(null);
    if (!lineId || Math.abs(hours) < 0.25) return;
    await onMoved(id, nextStart, nextEnd, lineId);
  }

  return (
    <div className="gantt-frame">
      <div className="gantt-scroll">
        <div className="gantt">
          <div className="gantt-head">
            <div className="line-meta">
              <code>LINE</code>
              <small>warp thread</small>
            </div>
            <div className="reed" aria-hidden="true">
              {days.map((d) => {
                const isToday = d.getTime() === today.getTime();
                return (
                  <span key={d.toISOString()} className={isToday ? "is-today" : undefined}>
                    {fmtDay(d.toISOString(), tz)}
                  </span>
                );
              })}
            </div>
          </div>
          {lines.map((line) => (
            <div className="gantt-row" key={line.id}>
              <div className="line-meta">
                <code translate="no">{line.code}</code>
                <small>
                  {line.name} · {line.kind} · {line.operators} ops
                </small>
              </div>
              <div
                className="warp"
                ref={(el) => {
                  trackRefs.current[line.id] = el;
                }}
              >
                {(blocksByLine.get(line.id) ?? []).map((b) => {
                  const left = xPct(drag?.id === b.id ? addHours(drag.originStart, drag.dxDays * 24) : b.starts_at);
                  const startIso = drag?.id === b.id ? addHours(drag.originStart, drag.dxDays * 24) : b.starts_at;
                  const endIso = drag?.id === b.id ? addHours(drag.originEnd, drag.dxDays * 24) : b.ends_at;
                  const w = widthPct(startIso, endIso);
                  const snag = b.has_conflict;
                  const gold = b.sop_state === "finishing" || b.sop_state === "packed";
                  return (
                    <button
                      key={b.id}
                      type="button"
                      className={[
                        "weft",
                        snag ? "is-snag" : "",
                        gold && !snag ? "is-gold" : "",
                        dim(b) ? "is-dim" : "",
                        drag?.id === b.id ? "is-dragging" : "",
                      ]
                        .filter(Boolean)
                        .join(" ")}
                      style={{ left: `${left}%`, width: `${Math.max(w, 1.2)}%` }}
                      aria-label={`${b.po_code} ${b.style_name} on ${b.line_code}`}
                      aria-pressed={selectedId === b.id}
                      onClick={() => onSelect(b)}
                      onPointerDown={(e) => {
                        if (!session.can_move) return;
                        (e.currentTarget as HTMLButtonElement).setPointerCapture(e.pointerId);
                        setDrag({
                          id: b.id,
                          originX: e.clientX,
                          dxDays: 0,
                          originStart: b.starts_at,
                          originEnd: b.ends_at,
                        });
                      }}
                      onPointerMove={(e) => {
                        if (!drag || drag.id !== b.id) return;
                        const track = trackRefs.current[line.id];
                        if (!track) return;
                        const rect = track.getBoundingClientRect();
                        const dayDelta = ((e.clientX - drag.originX) / rect.width) * DAYS;
                        setDrag({ ...drag, dxDays: Math.round(dayDelta * 48) / 48 });
                      }}
                      onPointerUp={() => {
                        void finishDrag();
                      }}
                    >
                      <span className="weft-label">
                        {b.buyer} · {b.style_name}
                      </span>
                      <span className="weft-meta">
                        {b.po_code} · {fmtPct(b.actual_pct)} act / {fmtPct(b.planned_pct)} plan
                      </span>
                    </button>
                  );
                })}
                {conflicts
                  .filter((c) => c.line_id === line.id && c.kind === "overlap")
                  .map((c) => (
                    <svg
                      key={c.id}
                      className="snag-knot"
                      style={{ left: `${xPct(c.starts_at)}%` }}
                      viewBox="0 0 40 40"
                      aria-hidden="true"
                    >
                      <path
                        d="M6 20c6-10 10 10 16 0s10 12 12 4M8 14c8 8 10-10 20 2M10 28c10-12 14 6 22-4"
                        fill="none"
                        stroke="#d4574a"
                        strokeWidth="2.2"
                        strokeLinecap="round"
                      />
                    </svg>
                  ))}
                {nowX >= 0 && nowX <= 100 ? (
                  <div className="shuttle" style={{ left: `calc(${nowX}% + var(--label-w) * 0)` }} />
                ) : null}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
