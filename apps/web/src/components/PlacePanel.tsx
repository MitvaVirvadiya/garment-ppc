"use client";

import { addHours } from "@/lib/format";
import type { Line, Order, Session } from "@/lib/types";
import { useEffect, useMemo, useState } from "react";

export function PlacePanel({
  session,
  orders,
  lines,
  onPlace,
  busy,
}: {
  session: Session;
  orders: Order[];
  lines: Line[];
  onPlace: (body: {
    order_id: string;
    line_id: string;
    starts_at: string;
    ends_at: string;
    planned_qty: number;
    sop_state: string;
  }) => Promise<void>;
  busy: boolean;
}) {
  const placeable = useMemo(
    () => orders.filter((o) => o.unplaced_qty > 0 || o.qty > 0),
    [orders],
  );
  const [orderId, setOrderId] = useState(placeable[0]?.id ?? "");
  const [lineId, setLineId] = useState(lines[0]?.id ?? "");
  const [startLocal, setStartLocal] = useState("");
  const [hours, setHours] = useState(16);
  const [qty, setQty] = useState(placeable[0]?.unplaced_qty || 1000);
  const [dirty, setDirty] = useState(false);

  useEffect(() => {
    const d = new Date();
    d.setDate(d.getDate() + 2);
    d.setHours(8, 0, 0, 0);
    const pad = (n: number) => String(n).padStart(2, "0");
    setStartLocal(
      `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`,
    );
  }, []);

  useEffect(() => {
    if (!orderId && placeable[0]) {
      setOrderId(placeable[0].id);
      setQty(Math.max(placeable[0].unplaced_qty, 200));
    }
  }, [placeable, orderId]);

  useEffect(() => {
    if (!lineId && lines[0]) setLineId(lines[0].id);
  }, [lines, lineId]);

  useEffect(() => {
    if (!dirty) return;
    const onLeave = (e: BeforeUnloadEvent) => {
      e.preventDefault();
      e.returnValue = "";
    };
    window.addEventListener("beforeunload", onLeave);
    return () => window.removeEventListener("beforeunload", onLeave);
  }, [dirty]);

  if (!session.can_place) {
    return (
      <section className="panel">
        <h2>Place on line</h2>
        <p className="muted">Anika’s viewer seat cannot load a line. Switch role to place.</p>
      </section>
    );
  }

  return (
    <section className="panel">
      <h2>Place on line</h2>
      <form
        className="row g-2"
        onSubmit={(e) => {
          e.preventDefault();
          if (!orderId || !lineId || !startLocal) return;
          const start = new Date(startLocal).toISOString();
          void onPlace({
            order_id: orderId,
            line_id: lineId,
            starts_at: start,
            ends_at: addHours(start, hours),
            planned_qty: qty,
            sop_state: "loaded",
          }).then(() => setDirty(false));
        }}
      >
        <div className="col-12 field">
          <label htmlFor="place-order">Order</label>
          <select
            id="place-order"
            className="form-select"
            name="order"
            autoComplete="off"
            value={orderId}
            onChange={(e) => {
              setOrderId(e.target.value);
              setDirty(true);
              const o = placeable.find((x) => x.id === e.target.value);
              if (o) setQty(Math.max(o.unplaced_qty, 200));
            }}
          >
            {placeable.map((o) => (
              <option key={o.id} value={o.id}>
                {o.po_code} · {o.style_name} · {o.unplaced_qty} open
              </option>
            ))}
          </select>
        </div>
        <div className="col-12 field">
          <label htmlFor="place-line">Line</label>
          <select
            id="place-line"
            className="form-select"
            name="line"
            autoComplete="off"
            value={lineId}
            onChange={(e) => {
              setLineId(e.target.value);
              setDirty(true);
            }}
          >
            {lines.map((l) => (
              <option key={l.id} value={l.id}>
                {l.code} · {l.name}
              </option>
            ))}
          </select>
        </div>
        <div className="col-md-7 field">
          <label htmlFor="place-start">Start</label>
          <input
            id="place-start"
            className="form-control"
            type="datetime-local"
            name="starts_at"
            autoComplete="off"
            value={startLocal}
            onChange={(e) => {
              setStartLocal(e.target.value);
              setDirty(true);
            }}
          />
        </div>
        <div className="col-md-5 field">
          <label htmlFor="place-hours">Hours</label>
          <input
            id="place-hours"
            className="form-control"
            type="number"
            name="hours"
            inputMode="numeric"
            min={2}
            max={72}
            value={hours}
            onChange={(e) => {
              setHours(Number(e.target.value));
              setDirty(true);
            }}
          />
        </div>
        <div className="col-12 field">
          <label htmlFor="place-qty">Planned pcs</label>
          <input
            id="place-qty"
            className="form-control"
            type="number"
            name="planned_qty"
            inputMode="numeric"
            min={1}
            value={qty}
            onChange={(e) => {
              setQty(Number(e.target.value));
              setDirty(true);
            }}
          />
        </div>
        <div className="col-12">
          <button className="btn btn-primary" type="submit" disabled={busy || !placeable.length}>
            {busy ? "Placing…" : "Place pick"}
          </button>
        </div>
      </form>
    </section>
  );
}
