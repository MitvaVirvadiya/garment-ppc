import { fmtPct } from "@/lib/format";
import type { Block } from "@/lib/types";

export function LiveBoard({ blocks }: { blocks: Block[] }) {
  const live = blocks
    .filter((b) => b.actual_pct > 0 && b.actual_pct < 100)
    .sort((a, b) => b.delay_minutes - a.delay_minutes)
    .slice(0, 6);

  return (
    <section className="panel">
      <h2>Live vs plan</h2>
      {live.length === 0 ? (
        <p className="muted">No picks are ticking. Loaded / sewing blocks will move every few seconds.</p>
      ) : (
        live.map((b) => {
          const late = b.delay_minutes >= 30;
          return (
            <div className="live-row" key={b.id}>
              <code translate="no">{b.line_code}</code>
              <div className={`meter ${late ? "late" : ""}`} title={b.style_name}>
                <i style={{ width: `${Math.min(b.planned_pct, 100)}%` }} />
                <b style={{ width: `${Math.min(b.actual_pct, 100)}%` }} />
              </div>
              <span className="muted" style={{ fontVariantNumeric: "tabular-nums" }}>
                {fmtPct(b.actual_pct)} / {fmtPct(b.planned_pct)}
              </span>
            </div>
          );
        })
      )}
    </section>
  );
}
