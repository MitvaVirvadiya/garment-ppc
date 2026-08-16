export type Role = "plant_head" | "merchandiser" | "viewer";

export type User = {
  id: string;
  slug: string;
  name: string;
  role: Role;
  plant_id: string | null;
  title: string;
};

export type Session = {
  user: User;
  can_place: boolean;
  can_move: boolean;
  can_advance: boolean;
  locked_plant_id: string | null;
  focus_buyer: string;
  kpi_set: "plant" | "merch" | "viewer" | string;
};

export type Plant = {
  id: string;
  code: string;
  name: string;
  city: string;
  region: string;
  timezone: string;
  shift_start: string;
  shift_end: string;
  shift_minutes: number;
};

export type Line = {
  id: string;
  unit_id: string;
  plant_id: string;
  plant_code: string;
  unit_code: string;
  code: string;
  name: string;
  kind: "cutting" | "sewing" | "finishing" | "wash" | string;
  operators: number;
  sam_minutes: number;
  shift_minutes: number;
};

export type Order = {
  id: string;
  po_code: string;
  style_code: string;
  style_name: string;
  buyer: string;
  plant_id: string;
  plant_code: string;
  qty: number;
  due_date: string;
  status: string;
  minutes_per_piece: number;
  colorway: string;
  merch_owner: string;
  placed_qty: number;
  unplaced_qty: number;
};

export type Block = {
  id: string;
  order_id: string;
  line_id: string;
  starts_at: string;
  ends_at: string;
  planned_qty: number;
  sop_state: string;
  pace: number;
  actual_pct: number;
  planned_pct: number;
  pieces_done: number;
  delay_minutes: number;
  notes: string;
  po_code: string;
  style_name: string;
  buyer: string;
  colorway: string;
  merch_owner: string;
  line_code: string;
  line_kind: string;
  plant_id: string;
  plant_code: string;
  conflict_ids: string[] | null;
  has_conflict: boolean;
};

export type Conflict = {
  id: string;
  kind: "overlap" | "over_shift" | string;
  line_id: string;
  line_code: string;
  plant_id: string;
  plant_code: string;
  block_ids: string[];
  starts_at: string;
  ends_at: string;
  minutes_over: number;
  message: string;
};

export type GanttPayload = {
  now: string;
  from: string;
  to: string;
  session: Session;
  plants: Plant[];
  lines: Line[];
  blocks: Block[];
  conflicts: Conflict[];
};

export type DashboardPayload = {
  session: Session;
  kpis: {
    collisions: number;
    lines: number;
    active_blocks: number;
    delayed_blocks: number;
    plan_attainment: number;
    minutes_over: number;
    packed_blocks: number;
    unplaced_qty: number;
    otd_risk: number;
    set: string;
  };
  conflicts: Conflict[];
  now: string;
};

export type SessionPayload = {
  session: Session;
  users: User[];
};

export const SOP_SEQUENCE = [
  "released",
  "marker",
  "cut",
  "loaded",
  "sewing",
  "mid_qc",
  "finishing",
  "packed",
  "ex_factory",
] as const;

export const SOP_LABEL: Record<string, string> = {
  released: "Released",
  marker: "Marker ready",
  cut: "Cut",
  loaded: "Loaded",
  sewing: "Sewing",
  mid_qc: "Mid-line QC",
  finishing: "Finishing",
  packed: "Packed",
  ex_factory: "Ex-factory",
};
