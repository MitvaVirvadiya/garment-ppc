package http

import (
	"context"
	"fmt"
	"time"

	"garment-ppc/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) ListPlants(ctx context.Context) ([]domain.Plant, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, code, name, city, region, timezone,
		       to_char(shift_start, 'HH24:MI'), to_char(shift_end, 'HH24:MI'), shift_minutes
		FROM org_plants
		ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Plant
	for rows.Next() {
		var p domain.Plant
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.City, &p.Region, &p.Timezone, &p.ShiftStart, &p.ShiftEnd, &p.ShiftMinutes); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) ListUnits(ctx context.Context, plantID string) ([]domain.Unit, error) {
	q := `
		SELECT id::text, plant_id::text, code, name
		FROM units`
	args := []any{}
	if plantID != "" {
		q += ` WHERE plant_id = $1`
		args = append(args, plantID)
	}
	q += ` ORDER BY code`
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Unit
	for rows.Next() {
		var u domain.Unit
		if err := rows.Scan(&u.ID, &u.PlantID, &u.Code, &u.Name); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (s *Store) ListLines(ctx context.Context, plantID string) ([]domain.Line, error) {
	q := `
		SELECT l.id::text, l.unit_id::text, p.id::text, p.code, u.code,
		       l.code, l.name, l.kind, l.operators, l.sam_minutes::float8, l.shift_minutes
		FROM lines l
		JOIN units u ON u.id = l.unit_id
		JOIN org_plants p ON p.id = u.plant_id`
	args := []any{}
	if plantID != "" {
		q += ` WHERE p.id = $1`
		args = append(args, plantID)
	}
	q += ` ORDER BY p.code, l.code`
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Line
	for rows.Next() {
		var l domain.Line
		if err := rows.Scan(&l.ID, &l.UnitID, &l.PlantID, &l.PlantCode, &l.UnitCode, &l.Code, &l.Name, &l.Kind, &l.Operators, &l.SAMMinutes, &l.ShiftMinutes); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (s *Store) LineByID(ctx context.Context, id string) (domain.Line, error) {
	var l domain.Line
	err := s.pool.QueryRow(ctx, `
		SELECT l.id::text, l.unit_id::text, p.id::text, p.code, u.code,
		       l.code, l.name, l.kind, l.operators, l.sam_minutes::float8, l.shift_minutes
		FROM lines l
		JOIN units u ON u.id = l.unit_id
		JOIN org_plants p ON p.id = u.plant_id
		WHERE l.id = $1`, id).Scan(
		&l.ID, &l.UnitID, &l.PlantID, &l.PlantCode, &l.UnitCode, &l.Code, &l.Name, &l.Kind, &l.Operators, &l.SAMMinutes, &l.ShiftMinutes,
	)
	return l, err
}

func (s *Store) ListUsers(ctx context.Context) ([]domain.User, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, slug, name, role, plant_id::text, title
		FROM users_roles
		ORDER BY role, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.User
	for rows.Next() {
		var u domain.User
		var plant *string
		if err := rows.Scan(&u.ID, &u.Slug, &u.Name, &u.Role, &plant, &u.Title); err != nil {
			return nil, err
		}
		if plant != nil && *plant != "" {
			u.PlantID = plant
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (s *Store) UserBySlug(ctx context.Context, slug string) (domain.User, error) {
	var u domain.User
	var plant *string
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, slug, name, role, plant_id::text, title
		FROM users_roles WHERE slug = $1`, slug).Scan(&u.ID, &u.Slug, &u.Name, &u.Role, &plant, &u.Title)
	if err != nil {
		return u, err
	}
	if plant != nil && *plant != "" {
		u.PlantID = plant
	}
	return u, nil
}

func (s *Store) ListOrders(ctx context.Context, plantID string) ([]domain.Order, error) {
	q := `
		SELECT o.id::text, o.po_code, o.style_code, o.style_name, o.buyer,
		       o.plant_id::text, p.code, o.qty, o.due_date::text, o.status,
		       o.minutes_per_piece::float8, o.colorway, o.merch_owner,
		       COALESCE((SELECT SUM(planned_qty) FROM gantt_blocks b WHERE b.order_id = o.id), 0)
		FROM plan_orders o
		JOIN org_plants p ON p.id = o.plant_id`
	args := []any{}
	if plantID != "" {
		q += ` WHERE o.plant_id = $1`
		args = append(args, plantID)
	}
	q += ` ORDER BY o.due_date, o.po_code`
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.POCode, &o.StyleCode, &o.StyleName, &o.Buyer, &o.PlantID, &o.PlantCode, &o.Qty, &o.DueDate, &o.Status, &o.MinutesPerPiece, &o.Colorway, &o.MerchOwner, &o.PlacedQty); err != nil {
			return nil, err
		}
		o.UnplacedQty = o.Qty - o.PlacedQty
		if o.UnplacedQty < 0 {
			o.UnplacedQty = 0
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (s *Store) ListBlocks(ctx context.Context, plantID string, from, to time.Time) ([]domain.Block, error) {
	q := `
		SELECT b.id::text, b.order_id::text, b.line_id::text, b.starts_at, b.ends_at,
		       b.planned_qty, b.sop_state, b.pace::float8, b.actual_pct::float8, b.pieces_done, b.notes,
		       o.po_code, o.style_name, o.buyer, o.colorway, o.merch_owner,
		       l.code, l.kind, p.id::text, p.code
		FROM gantt_blocks b
		JOIN plan_orders o ON o.id = b.order_id
		JOIN lines l ON l.id = b.line_id
		JOIN units u ON u.id = l.unit_id
		JOIN org_plants p ON p.id = u.plant_id
		WHERE b.starts_at < $1 AND b.ends_at > $2`
	args := []any{to, from}
	if plantID != "" {
		q += ` AND p.id = $3`
		args = append(args, plantID)
	}
	q += ` ORDER BY p.code, l.code, b.starts_at`
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	now := time.Now()
	var out []domain.Block
	for rows.Next() {
		var b domain.Block
		if err := rows.Scan(&b.ID, &b.OrderID, &b.LineID, &b.StartsAt, &b.EndsAt, &b.PlannedQty, &b.SOPState, &b.Pace, &b.ActualPct, &b.PiecesDone, &b.Notes, &b.POCode, &b.StyleName, &b.Buyer, &b.Colorway, &b.MerchOwner, &b.LineCode, &b.LineKind, &b.PlantID, &b.PlantCode); err != nil {
			return nil, err
		}
		b.PlannedPct = domain.PlannedPct(now, b.StartsAt, b.EndsAt)
		b.DelayMinutes = domain.DelayMinutes(b.PlannedPct, b.ActualPct, b.StartsAt, b.EndsAt)
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *Store) BlockByID(ctx context.Context, id string) (domain.Block, error) {
	var b domain.Block
	err := s.pool.QueryRow(ctx, `
		SELECT b.id::text, b.order_id::text, b.line_id::text, b.starts_at, b.ends_at,
		       b.planned_qty, b.sop_state, b.pace::float8, b.actual_pct::float8, b.pieces_done, b.notes,
		       o.po_code, o.style_name, o.buyer, o.colorway, o.merch_owner,
		       l.code, l.kind, p.id::text, p.code
		FROM gantt_blocks b
		JOIN plan_orders o ON o.id = b.order_id
		JOIN lines l ON l.id = b.line_id
		JOIN units u ON u.id = l.unit_id
		JOIN org_plants p ON p.id = u.plant_id
		WHERE b.id = $1`, id).Scan(
		&b.ID, &b.OrderID, &b.LineID, &b.StartsAt, &b.EndsAt, &b.PlannedQty, &b.SOPState, &b.Pace, &b.ActualPct, &b.PiecesDone, &b.Notes,
		&b.POCode, &b.StyleName, &b.Buyer, &b.Colorway, &b.MerchOwner, &b.LineCode, &b.LineKind, &b.PlantID, &b.PlantCode,
	)
	if err != nil {
		return b, err
	}
	now := time.Now()
	b.PlannedPct = domain.PlannedPct(now, b.StartsAt, b.EndsAt)
	b.DelayMinutes = domain.DelayMinutes(b.PlannedPct, b.ActualPct, b.StartsAt, b.EndsAt)
	return b, nil
}

func (s *Store) InsertBlock(ctx context.Context, orderID, lineID string, start, end time.Time, qty int, sop, notes string, pace float64) (string, error) {
	if pace <= 0 {
		pace = 1
	}
	if sop == "" {
		sop = "released"
	}
	var id string
	err := s.pool.QueryRow(ctx, `
		INSERT INTO gantt_blocks (order_id, line_id, starts_at, ends_at, planned_qty, sop_state, pace, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id::text`, orderID, lineID, start, end, qty, sop, pace, notes).Scan(&id)
	if err != nil {
		return "", err
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO sop_states (block_id, state, actor) VALUES ($1,$2,$3)`, id, sop, "planner")
	return id, err
}

func (s *Store) UpdateBlock(ctx context.Context, id, lineID string, start, end time.Time, qty *int) error {
	cmd, err := s.pool.Exec(ctx, `
		UPDATE gantt_blocks
		SET line_id = COALESCE(NULLIF($2,'')::uuid, line_id),
		    starts_at = $3,
		    ends_at = $4,
		    planned_qty = COALESCE($5, planned_qty)
		WHERE id = $1`, id, lineID, start, end, qty)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) DeleteBlock(ctx context.Context, id string) error {
	cmd, err := s.pool.Exec(ctx, `DELETE FROM gantt_blocks WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) AdvanceSOP(ctx context.Context, id, actor string, target string) (string, error) {
	block, err := s.BlockByID(ctx, id)
	if err != nil {
		return "", err
	}
	next := target
	if next == "" {
		var ok bool
		next, ok = domain.NextSOP(block.SOPState)
		if !ok {
			return block.SOPState, fmt.Errorf("already at last SOP gate")
		}
	}
	if domain.SOPIndex(next) < 0 {
		return "", fmt.Errorf("unknown SOP state %q", next)
	}
	_, err = s.pool.Exec(ctx, `UPDATE gantt_blocks SET sop_state = $2 WHERE id = $1`, id, next)
	if err != nil {
		return "", err
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO sop_states (block_id, state, actor) VALUES ($1,$2,$3)`, id, next, actor)
	return next, err
}

func (s *Store) SOPHistory(ctx context.Context, blockID string) ([]domain.SOPEntry, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, block_id::text, state, entered_at, actor
		FROM sop_states WHERE block_id = $1 ORDER BY entered_at`, blockID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.SOPEntry
	for rows.Next() {
		var e domain.SOPEntry
		if err := rows.Scan(&e.ID, &e.BlockID, &e.State, &e.EnteredAt, &e.Actor); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) RecentTicks(ctx context.Context, blockID string, limit int) ([]domain.Tick, error) {
	if limit <= 0 || limit > 80 {
		limit = 24
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, block_id::text, actual_pct::float8, planned_pct::float8, pieces_done, delay_minutes, ticked_at
		FROM progress_ticks
		WHERE ($1 = '' OR block_id = $1::uuid)
		ORDER BY ticked_at DESC
		LIMIT $2`, blockID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Tick
	for rows.Next() {
		var t domain.Tick
		if err := rows.Scan(&t.ID, &t.BlockID, &t.ActualPct, &t.PlannedPct, &t.PiecesDone, &t.DelayMinutes, &t.TickedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) PlantByCode(ctx context.Context, code string) (domain.Plant, error) {
	var p domain.Plant
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, code, name, city, region, timezone,
		       to_char(shift_start, 'HH24:MI'), to_char(shift_end, 'HH24:MI'), shift_minutes
		FROM org_plants WHERE code = $1`, code).Scan(
		&p.ID, &p.Code, &p.Name, &p.City, &p.Region, &p.Timezone, &p.ShiftStart, &p.ShiftEnd, &p.ShiftMinutes,
	)
	return p, err
}

func (s *Store) HasAnyPlant(ctx context.Context) (bool, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM org_plants`).Scan(&n)
	return n > 0, err
}

func IsNoRows(err error) bool {
	return err == pgx.ErrNoRows
}
