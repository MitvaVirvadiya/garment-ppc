# Loom · Garment PPC

**Thesis:** Multi-plant PPC as a loom — warp is the line, time is the shuttle, collisions glow as a snag.

**Stack:** Next.js 15 · Tabler (restyled) · custom warp/weft Gantt · Go Fiber · pgx · Postgres via `DATABASE_URL`

**Shots**
1. Tirupur warp with a red knot on TK-S3 (H&M × Next).
2. Same frame after the pick is slid — snag count drops; gold shuttle at *now*.
3. Live-vs-plan meters ticking; Anika’s Gazipur-only viewer seat.

**Resume**
- Built a multi-plant garment PPC system with line-level Gantt scheduling and capacity conflict detection.
- Simulated shop-floor progress feeds to highlight live-versus-plan delays.

**Run:** `docker compose up -d && cd apps/api && go run ./cmd/seed && go run ./cmd/api` then `cd apps/web && npm run dev` → http://localhost:3014

**2-minute demo:** Kavya / Tirupur → red snag on TK-S3 → drag the Next blouse off the warp → watch Primark % tick late → switch to Anika (Gazipur, read-only).
