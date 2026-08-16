# Loom — Multi-plant Garment PPC

Operational control layer for a three-plant knit / woven / denim house. A plant PPC head opens the warp, sees this week’s collisions, slides a pick, and watches live % tick against plan.

This is **not** project 13 (atelier planning + SOP gates before commit). Loom is **loading lines across plants** after the plan exists.

---

## Visual identity — Loom control room

Named before any component was scaffolded.

| Token | Value | Role |
|---|---|---|
| Loom dark | `#0F1412` | Mill-floor night, oil + iron |
| Warp | `#E7DCC8` | Cream warp threads |
| Weft teal | `#4B8F8C` | Loaded picks |
| Collision | `#D4574A` | Snag / knot |
| Gold pick | `#F0C36A` | Shuttle (now) + finishing |

**Type:** Barlow Condensed (line IDs, KPIs) + Barlow (body). No Inter.

**Layout thesis:** The first surface is the warp, not a KPI card grid. Plant rail and role seat sit in the header like loom-frame labels. Gantt rows are taut warp threads; work is weft woven through them; time is a gold shuttle. Overlaps glow as a knotted snag.

**Signature (one risk):** Warp/weft Gantt. Tabler is remapped to this mill; the default Tabler dashboard wallpaper is gone.

---

## Stack

| Layer | Choice |
|---|---|
| Frontend | Next.js 15 (App Router) |
| Components | Tabler (`@tabler/core` + `@tabler/icons-react`), restyled |
| Gantt | Custom CSS + SVG warp/weft (not MS Project) |
| Backend | Go Fiber |
| DB | PostgreSQL via **pgx** · `DATABASE_URL` only |
| Auth (demo) | Role switch: plant head / merchandiser / viewer |

---

## Architecture

```
                 ┌──────────────────────────┐
  Seat / plant   │  Next.js 15 :3014        │
  query params   │  Tabler + Loom theme     │
                 │  WarpGantt  ·  SOP tray  │
                 └────────────┬─────────────┘
                              │ REST  X-Demo-User
                 ┌────────────▼─────────────┐
                 │  Go Fiber :8014          │
                 │  RBAC · conflicts · SOP  │
                 │  sim ticker every 4s     │
                 └────────────┬─────────────┘
                              │ pgx
                 ┌────────────▼─────────────┐
                 │  Postgres                │
                 │  plants → units → lines  │
                 │  orders → blocks → ticks │
                 └──────────────────────────┘
```

Point this at Neon / Supabase / local by changing `DATABASE_URL` only. `sslmode` lives in the URL. There is no `if neon` / `if supabase` branch.

---

## Data

`org_plants` · `units` · `lines` · `plan_orders` · `gantt_blocks` · `sop_states` · `progress_ticks` · `users_roles`

Seed (idempotent truncate + insert):

- **Tirupur Knitworks** (TK) — 4 lines
- **Noida Wovens** (NW) — 3 lines
- **Gazipur Denim** (GD) — 3 lines
- 15 export POs (H&M, Next, Primark, M&S, Levi’s, Decathlon…)
- 2-week window relative to today
- **3 intentional snags:** TK-S3 overlap, GD-S2 overlap (includes today), NW-F1 minutes > shift

Seats:

| Slug | Person | Role | Scope |
|---|---|---|---|
| `kavya` | Kavya Menon | Plant head | Tirupur only, can move / place / SOP |
| `rohan` | Rohan Desai | Merchandiser | All plants, his POs highlighted |
| `anika` | Anika Rahman | Viewer | Gazipur read-only |

---

## Run

```bash
cd full-stack/garment-ppc
cp .env.example .env
cp apps/web/.env.example apps/web/.env.local

docker compose -f ../docker-compose.yml --project-directory .. up -d   # shared Postgres :5432
cd apps/api && go run ./cmd/seed && go run ./cmd/api

# other terminal
cd apps/web && npm install && npm run dev
```

Open [http://localhost:3014](http://localhost:3014)

API health: [http://localhost:8014/health](http://localhost:8014/health)

Or `make up && make seed` then `make api` / `make web`.

### DATABASE_URL examples

```env
# Required
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/garment_ppc

# Local Docker / Homebrew
# DATABASE_URL=postgresql://postgres:postgres@localhost:5432/garment_ppc

# Neon
# DATABASE_URL=postgresql://USER:PASSWORD@ep-xxx.region.aws.neon.tech/neondb?sslmode=require

# Supabase (direct or pooler)
# DATABASE_URL=postgresql://postgres:PASSWORD@db.<ref>.supabase.co:5432/postgres?sslmode=require
# DATABASE_URL=postgresql://postgres.<ref>:PASSWORD@aws-0-<region>.pooler.supabase.com:6543/postgres?sslmode=require
```

This compose maps **54314 → 5432** so it does not collide with another local Postgres.

---

## 2-minute demo

1. Open as **Kavya / Tirupur**. The warp shows a red snag on **TK-S3** (H&M polo vs Next blouse).
2. Drag the Next blouse weft off the collision (or place it a day later). The knot clears; snag count drops.
3. Watch **TK-S1** (Primark jersey) — actual % ticks every 4s against the paler plan bar. It is seeded slow on purpose.
4. Switch seat to **Anika**. Plant locks to Gazipur, Place is disabled, only GD lines remain.

Optional: **Rohan** dims blocks he does not own. **NW-F1** on day +4 is the over-shift snag.

---

## Resume bullets

- Built a multi-plant garment PPC system with line-level Gantt scheduling and capacity conflict detection.
- Added SOP-governed execution states and role-scoped dashboards across plants and units.
- Simulated shop-floor progress feeds to highlight live-versus-plan delays.

---

## Interview talking points

- Conflict engine is pure functions over intervals: overlap on a line, plus minutes booked vs shift length in the plant timezone. The API never blocks a place — PPC needs to *see* the snag.
- Demo clock is relative (`today − 3` … `today + 11`) so the Loom always has a collision this week.
- Role is a header (`X-Demo-User`), not a fake JWT: plant head is fenced to one mill; merch is cross-plant; viewer is read-only.
- Gantt is CSS warp threads + weft buttons + SVG knots, not a licensed scheduler.

---

## API (short)

| Method | Path | Notes |
|---|---|---|
| GET | `/health` | db + sim status |
| GET | `/api/session` | current seat + all users |
| GET | `/api/gantt?plant=TK` | lines, blocks, conflicts |
| GET | `/api/dashboard` | role KPI set |
| POST | `/api/blocks` | place |
| PATCH | `/api/blocks/:id` | move |
| POST | `/api/blocks/:id/sop` | `{ "action": "advance" }` |
| POST | `/api/sim/tick` | force a tick |

Header: `X-Demo-User: kavya|rohan|anika`
