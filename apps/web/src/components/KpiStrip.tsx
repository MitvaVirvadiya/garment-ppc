import { fmtPct, fmtQty } from "@/lib/format";
import type { DashboardPayload, Session } from "@/lib/types";

export function KpiStrip({ dash, session }: { dash: DashboardPayload; session: Session }) {
  const k = dash.kpis;
  if (session.kpi_set === "merch") {
    return (
      <div className="kpi-strip">
        <div className="kpi gold">
          <em>OTD risk</em>
          <b>{k.otd_risk}</b>
        </div>
        <div className="kpi">
          <em>Unplaced pcs</em>
          <b>{fmtQty(k.unplaced_qty)}</b>
        </div>
        <div className={`kpi ${k.collisions ? "alert" : ""}`}>
          <em>Snags on your POs</em>
          <b>{k.collisions}</b>
        </div>
        <div className="kpi">
          <em>Plan attainment</em>
          <b>{fmtPct(k.plan_attainment)}</b>
        </div>
      </div>
    );
  }
  if (session.kpi_set === "viewer") {
    return (
      <div className="kpi-strip">
        <div className="kpi">
          <em>Active picks</em>
          <b>{k.active_blocks}</b>
        </div>
        <div className={`kpi ${k.delayed_blocks ? "alert" : ""}`}>
          <em>Late vs plan</em>
          <b>{k.delayed_blocks}</b>
        </div>
        <div className="kpi">
          <em>Packed</em>
          <b>{k.packed_blocks}</b>
        </div>
        <div className="kpi gold">
          <em>Attainment</em>
          <b>{fmtPct(k.plan_attainment)}</b>
        </div>
      </div>
    );
  }
  return (
    <div className="kpi-strip">
      <div className={`kpi ${k.collisions ? "alert" : ""}`}>
        <em>Snags this window</em>
        <b>{k.collisions}</b>
      </div>
      <div className={`kpi ${k.delayed_blocks ? "alert" : ""}`}>
        <em>Late lines</em>
        <b>{k.delayed_blocks}</b>
      </div>
      <div className="kpi gold">
        <em>Minutes over shift</em>
        <b>{fmtQty(k.minutes_over)}</b>
      </div>
      <div className="kpi">
        <em>Plan attainment</em>
        <b>{fmtPct(k.plan_attainment)}</b>
      </div>
    </div>
  );
}
