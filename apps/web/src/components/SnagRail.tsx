import { fmtWhen } from "@/lib/format";
import type { Conflict } from "@/lib/types";
import { IconAlertTriangle } from "@tabler/icons-react";

export function SnagRail({
  conflicts,
  activeId,
  onFocus,
  tz,
}: {
  conflicts: Conflict[];
  activeId: string | null;
  onFocus: (c: Conflict) => void;
  tz: string;
}) {
  return (
    <div className="snag-rail">
      <div className="snag-count" aria-live="polite">
        <strong>{conflicts.length}</strong>
        <span>{conflicts.length === 1 ? "snag" : "snags"}</span>
      </div>
      <div className="snag-scroller">
        {conflicts.length === 0 ? (
          <div className="banner">Warp is clean across this window. Place more work to test capacity.</div>
        ) : (
          conflicts.map((c) => (
            <button
              key={c.id}
              type="button"
              className={`snag-chip ${activeId === c.id ? "is-active" : ""}`}
              onClick={() => onFocus(c)}
            >
              <b>
                <IconAlertTriangle size={14} aria-hidden="true" /> {c.plant_code} · {c.line_code} ·{" "}
                {c.kind === "overlap" ? "overlap" : "over shift"}
              </b>
              <small>
                {c.message} · {fmtWhen(c.starts_at, tz)}
              </small>
            </button>
          ))
        )}
      </div>
    </div>
  );
}
