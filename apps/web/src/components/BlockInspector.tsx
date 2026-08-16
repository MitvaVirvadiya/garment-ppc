"use client";

import { fmtPct, fmtQty, fmtWhen } from "@/lib/format";
import { SOP_LABEL, SOP_SEQUENCE, type Block, type Session } from "@/lib/types";

export function BlockInspector({
  block,
  session,
  tz,
  onAdvance,
  busy,
}: {
  block: Block | null;
  session: Session;
  tz: string;
  onAdvance: () => void;
  busy: boolean;
}) {
  if (!block) {
    return (
      <section className="panel">
        <h2>Pick inspector</h2>
        <p className="muted">Select a weft on the warp to read SOP, live % and delay.</p>
      </section>
    );
  }
  const idx = SOP_SEQUENCE.indexOf(block.sop_state as (typeof SOP_SEQUENCE)[number]);
  const late = block.delay_minutes >= 30 && block.actual_pct < 100;
  return (
    <section className="panel">
      <h2>{block.po_code}</h2>
      <p>
        <strong>
          {block.buyer} · {block.style_name}
        </strong>
        <br />
        <span className="muted">
          {block.colorway} · {fmtQty(block.planned_qty)} pcs · {block.line_code}
        </span>
      </p>
      <p className="muted">
        {fmtWhen(block.starts_at, tz)} → {fmtWhen(block.ends_at, tz)}
      </p>
      <div className="live-row" style={{ gridTemplateColumns: "1fr" }}>
        <div>
          <div className="muted">
            Live {fmtPct(block.actual_pct)} vs plan {fmtPct(block.planned_pct)}
            {late ? ` · ${block.delay_minutes} min late` : ""}
          </div>
          <div className={`meter ${late ? "late" : ""}`} aria-hidden="true">
            <i style={{ width: `${Math.min(block.planned_pct, 100)}%` }} />
            <b style={{ width: `${Math.min(block.actual_pct, 100)}%` }} />
          </div>
        </div>
      </div>
      <div className="sop-steps" aria-label="SOP gates">
        {SOP_SEQUENCE.map((s, i) => (
          <span key={s} className={i < idx ? "done" : i === idx ? "now" : undefined}>
            {SOP_LABEL[s]}
          </span>
        ))}
      </div>
      {block.notes ? <p className="muted mt-2">{block.notes}</p> : null}
      {block.has_conflict ? (
        <p className="mt-2" style={{ color: "var(--snag)" }}>
          This pick is in a snag. Slide it off the colliding warp to clear the knot.
        </p>
      ) : null}
      <div className="mt-3">
        <button
          className="btn btn-primary btn-sm"
          type="button"
          disabled={!session.can_advance || busy || block.sop_state === "ex_factory"}
          onClick={onAdvance}
        >
          {busy ? "Advancing…" : "Advance SOP"}
        </button>
        {!session.can_advance ? <span className="muted ms-2">View-only role</span> : null}
      </div>
    </section>
  );
}
