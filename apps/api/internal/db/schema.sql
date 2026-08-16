CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS org_plants (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code            TEXT UNIQUE NOT NULL,
    name            TEXT NOT NULL,
    city            TEXT NOT NULL,
    region          TEXT NOT NULL,
    timezone        TEXT NOT NULL,
    shift_start     TIME NOT NULL DEFAULT '08:00',
    shift_end       TIME NOT NULL DEFAULT '17:00',
    shift_minutes   INT NOT NULL DEFAULT 480
);

CREATE TABLE IF NOT EXISTS units (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plant_id    UUID NOT NULL REFERENCES org_plants(id) ON DELETE CASCADE,
    code        TEXT NOT NULL,
    name        TEXT NOT NULL,
    UNIQUE (plant_id, code)
);

CREATE TABLE IF NOT EXISTS lines (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    unit_id         UUID NOT NULL REFERENCES units(id) ON DELETE CASCADE,
    code            TEXT NOT NULL,
    name            TEXT NOT NULL,
    kind            TEXT NOT NULL CHECK (kind IN ('cutting', 'sewing', 'finishing', 'wash')),
    operators       INT NOT NULL DEFAULT 24,
    sam_minutes     NUMERIC(6, 2) NOT NULL DEFAULT 12,
    shift_minutes   INT NOT NULL DEFAULT 480,
    UNIQUE (unit_id, code)
);

CREATE TABLE IF NOT EXISTS users_roles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        TEXT UNIQUE NOT NULL,
    name        TEXT NOT NULL,
    role        TEXT NOT NULL CHECK (role IN ('plant_head', 'merchandiser', 'viewer')),
    plant_id    UUID REFERENCES org_plants(id),
    title       TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS plan_orders (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    po_code             TEXT UNIQUE NOT NULL,
    style_code          TEXT NOT NULL,
    style_name          TEXT NOT NULL,
    buyer               TEXT NOT NULL,
    plant_id            UUID NOT NULL REFERENCES org_plants(id),
    qty                 INT NOT NULL,
    due_date            DATE NOT NULL,
    status              TEXT NOT NULL DEFAULT 'open',
    minutes_per_piece   NUMERIC(6, 2) NOT NULL,
    colorway            TEXT NOT NULL,
    merch_owner         TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS gantt_blocks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID NOT NULL REFERENCES plan_orders(id) ON DELETE CASCADE,
    line_id         UUID NOT NULL REFERENCES lines(id) ON DELETE CASCADE,
    starts_at       TIMESTAMPTZ NOT NULL,
    ends_at         TIMESTAMPTZ NOT NULL,
    planned_qty     INT NOT NULL,
    sop_state       TEXT NOT NULL DEFAULT 'released',
    pace            NUMERIC(4, 2) NOT NULL DEFAULT 1.00,
    actual_pct      NUMERIC(5, 2) NOT NULL DEFAULT 0,
    pieces_done     INT NOT NULL DEFAULT 0,
    notes           TEXT NOT NULL DEFAULT '',
    CHECK (ends_at > starts_at)
);

CREATE TABLE IF NOT EXISTS sop_states (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    block_id    UUID NOT NULL REFERENCES gantt_blocks(id) ON DELETE CASCADE,
    state       TEXT NOT NULL,
    entered_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    actor       TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS progress_ticks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    block_id        UUID NOT NULL REFERENCES gantt_blocks(id) ON DELETE CASCADE,
    actual_pct      NUMERIC(5, 2) NOT NULL,
    planned_pct     NUMERIC(5, 2) NOT NULL,
    pieces_done     INT NOT NULL,
    delay_minutes   INT NOT NULL DEFAULT 0,
    ticked_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_units_plant ON units (plant_id);
CREATE INDEX IF NOT EXISTS idx_lines_unit ON lines (unit_id);
CREATE INDEX IF NOT EXISTS idx_orders_plant ON plan_orders (plant_id);
CREATE INDEX IF NOT EXISTS idx_blocks_line_time ON gantt_blocks (line_id, starts_at, ends_at);
CREATE INDEX IF NOT EXISTS idx_blocks_order ON gantt_blocks (order_id);
CREATE INDEX IF NOT EXISTS idx_sop_block ON sop_states (block_id, entered_at);
CREATE INDEX IF NOT EXISTS idx_ticks_block ON progress_ticks (block_id, ticked_at DESC);
