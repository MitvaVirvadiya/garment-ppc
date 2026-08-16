package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"garment-ppc/internal/config"
	"garment-ppc/internal/db"
	"garment-ppc/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env", "../../.env")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatal(err)
	}
	if err := seed(ctx, pool); err != nil {
		log.Fatal(err)
	}
	fmt.Println("seeded garment_ppc — 3 plants, 10 lines, 2-week warp, 3 snags")
}

func seed(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `
		TRUNCATE progress_ticks, sop_states, gantt_blocks, plan_orders, users_roles, lines, units, org_plants CASCADE`); err != nil {
		return err
	}

	ist, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return err
	}
	bst, err := time.LoadLocation("Asia/Dhaka")
	if err != nil {
		return err
	}

	must := func(id string, e error) string {
		if e != nil {
			panic(e)
		}
		return id
	}

	tk := must(insertPlant(ctx, pool, "TK", "Tirupur Knitworks", "Tirupur", "Tamil Nadu", "Asia/Kolkata", "08:00", "17:00", 480))
	nw := must(insertPlant(ctx, pool, "NW", "Noida Wovens", "Noida", "Uttar Pradesh", "Asia/Kolkata", "08:00", "18:00", 540))
	gd := must(insertPlant(ctx, pool, "GD", "Gazipur Denim", "Gazipur", "Dhaka Division", "Asia/Dhaka", "08:00", "17:00", 480))

	tkCut := must(insertUnit(ctx, pool, tk, "CUT", "Cutting Hall"))
	tkSew := must(insertUnit(ctx, pool, tk, "SEW", "Knit Sewing"))
	nwCut := must(insertUnit(ctx, pool, nw, "CUT", "Cut & Sew Prep"))
	nwFin := must(insertUnit(ctx, pool, nw, "FIN", "Finishing Bay"))
	gdSew := must(insertUnit(ctx, pool, gd, "SEW", "Denim Sewing"))
	gdWash := must(insertUnit(ctx, pool, gd, "WASH", "Wash House"))

	tkC1 := must(insertLine(ctx, pool, tkCut, "TK-C1", "Band-knife 1", "cutting", 8, 3.2, 480))
	tkS1 := must(insertLine(ctx, pool, tkSew, "TK-S1", "Pique line 14", "sewing", 32, 14.5, 480))
	tkS2 := must(insertLine(ctx, pool, tkSew, "TK-S2", "Jersey line 09", "sewing", 28, 9.8, 480))
	tkS3 := must(insertLine(ctx, pool, tkSew, "TK-S3", "Fashion knits 03", "sewing", 30, 16.2, 480))
	nwC1 := must(insertLine(ctx, pool, nwCut, "NW-C1", "Spreader 2", "cutting", 10, 4.1, 540))
	nwS1 := must(insertLine(ctx, pool, nwCut, "NW-S1", "Woven sew 06", "sewing", 36, 22.0, 540))
	nwF1 := must(insertLine(ctx, pool, nwFin, "NW-F1", "Press & pack", "finishing", 18, 6.4, 540))
	gdS1 := must(insertLine(ctx, pool, gdSew, "GD-S1", "5-pocket A", "sewing", 40, 28.0, 480))
	gdS2 := must(insertLine(ctx, pool, gdSew, "GD-S2", "5-pocket B", "sewing", 40, 26.5, 480))
	gdW1 := must(insertLine(ctx, pool, gdWash, "GD-W1", "Stone wash 1", "wash", 12, 18.0, 480))

	if _, err := pool.Exec(ctx, `
		INSERT INTO users_roles (slug, name, role, plant_id, title) VALUES
		('kavya', 'Kavya Menon', 'plant_head', $1, 'Plant Head — Tirupur Knitworks'),
		('rohan', 'Rohan Desai', 'merchandiser', NULL, 'Group merchandiser — EU knit + denim'),
		('anika', 'Anika Rahman', 'viewer', $2, 'PPC viewer — Gazipur Denim')`, tk, gd); err != nil {
		return err
	}

	type ord struct {
		po, style, name, buyer, plant, color, merch string
		qty                                         int
		due                                         int
		sam                                         float64
	}
	orders := []ord{
		{"TK-PO-4412", "HM-PK-4412", "Men's pique polo", "H&M", tk, "Deep navy", "Rohan Desai", 12000, 12, 14.5},
		{"TK-PO-4418", "NX-VB-4418", "Women's viscose blouse", "Next PLC", tk, "Sand dune", "Rohan Desai", 8400, 13, 16.2},
		{"TK-PO-4421", "PR-JT-4421", "Kids jersey tee 3-pack", "Primark", tk, "Bright white", "Meera Iyer", 22000, 10, 9.8},
		{"TK-PO-4425", "CA-RH-4425", "Rib henley", "C&A", tk, "Heather grey", "Meera Iyer", 6000, 14, 12.4},
		{"TK-PO-4430", "TG-BR-4430", "Baby romper", "Target", tk, "Cloud mint", "Rohan Desai", 15000, 15, 11.0},
		{"TK-PO-4433", "LS-LQ-4433", "Lounge short", "Lifestyle", tk, "Ink black", "Meera Iyer", 9600, 16, 8.6},
		{"NW-PO-1188", "MS-CH-1188", "Cotton chino", "Marks & Spencer", nw, "Khaki", "Rohan Desai", 4800, 11, 22.0},
		{"NW-PO-1192", "ZR-LN-1192", "Linen camp shirt", "Zara", nw, "Ecru", "Priya Shah", 3200, 12, 18.5},
		{"NW-PO-1199", "UQ-OX-1199", "Oxford button-down", "Uniqlo", nw, "Sky", "Priya Shah", 5600, 14, 19.0},
		{"NW-PO-1204", "WS-KT-1204", "Handloom kurta", "Westside", nw, "Indigo vat", "Rohan Desai", 2400, 16, 24.0},
		{"NW-PO-1210", "MS-AG-1210", "Autograph wrap dress", "Marks & Spencer", nw, "Claret", "Priya Shah", 1800, 17, 31.0},
		{"GD-PO-7701", "LV-5P-7701", "Classic 5-pocket", "Levi's", gd, "Rigid indigo", "Rohan Desai", 9000, 13, 28.0},
		{"GD-PO-7704", "HM-MJ-7704", "Mom jean", "H&M", gd, "Vintage wash", "Rohan Desai", 7200, 14, 26.5},
		{"GD-PO-7708", "DC-HS-7708", "Hiking short", "Decathlon", gd, "Olive drab", "Anika Rahman", 11000, 11, 17.4},
		{"GD-PO-7712", "MG-WL-7712", "Wide-leg crop", "Mango", gd, "Ecru denim", "Anika Rahman", 4400, 16, 27.2},
	}
	oids := map[string]string{}
	for _, o := range orders {
		due := time.Now().In(ist).AddDate(0, 0, o.due).Format("2006-01-02")
		id, err := insertOrder(ctx, pool, o.po, o.style, o.name, o.buyer, o.plant, o.qty, due, o.sam, o.color, o.merch)
		if err != nil {
			return err
		}
		oids[o.po] = id
	}

	at := func(loc *time.Location, day, hour, min int) time.Time {
		now := time.Now().In(loc)
		return time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, loc).AddDate(0, 0, day)
	}

	type blk struct {
		po, line, sop, notes string
		d0, h0, m0, d1, h1, m1 int
		qty                    int
		pace                   float64
		loc                    *time.Location
	}
	blocks := []blk{
		// Tirupur cutting
		{"TK-PO-4421", tkC1, "packed", "Kids tee lay complete", -3, 8, 0, -1, 16, 30, 22000, 1.0, ist},
		{"TK-PO-4430", tkC1, "cut", "Romper markers on table 4", 0, 8, 0, 2, 16, 0, 15000, 0.95, ist},
		{"TK-PO-4433", tkC1, "released", "Lounge short — wait for fabric", 5, 8, 0, 7, 15, 0, 9600, 1.0, ist},
		// Tirupur sewing — live delayed
		{"TK-PO-4421", tkS1, "sewing", "Needle heat on white jersey", -1, 8, 0, 6, 17, 0, 22000, 0.68, ist},
		{"TK-PO-4425", tkS2, "mid_qc", "Henley placket attach", -2, 8, 0, 2, 17, 0, 6000, 1.05, ist},
		{"TK-PO-4433", tkS2, "released", "Lounge short sew window", 7, 8, 0, 11, 17, 0, 9600, 1.0, ist},
		// SNAG 1 — TK-S3 overlap
		{"TK-PO-4412", tkS3, "loaded", "H&M polo loaded on fashion knits", 2, 8, 0, 4, 17, 0, 7000, 1.0, ist},
		{"TK-PO-4418", tkS3, "released", "Next blouse parked on same warp", 3, 8, 0, 5, 13, 0, 8400, 1.0, ist},
		{"TK-PO-4430", tkS3, "released", "Romper sew after collision cleared", 8, 8, 0, 11, 17, 0, 15000, 1.0, ist},
		// Noida
		{"NW-PO-1188", nwC1, "cut", "Chino marker 32–38", -3, 8, 0, -2, 17, 0, 4800, 1.0, ist},
		{"NW-PO-1192", nwC1, "cut", "Linen camp — shade lot B", -1, 8, 0, 0, 16, 0, 3200, 1.0, ist},
		{"NW-PO-1199", nwC1, "loaded", "Oxford spread tonight", 1, 8, 0, 2, 17, 0, 5600, 1.0, ist},
		{"NW-PO-1204", nwC1, "released", "Kurta handloom — wait shade", 6, 8, 0, 7, 17, 0, 2400, 1.0, ist},
		{"NW-PO-1188", nwS1, "sewing", "Chino side-seam", -1, 8, 0, 3, 18, 0, 4800, 0.78, ist},
		{"NW-PO-1192", nwS1, "loaded", "Linen — no overlap, follows chino", 4, 8, 0, 6, 18, 0, 3200, 1.0, ist},
		{"NW-PO-1199", nwS1, "released", "Oxford sew", 7, 8, 0, 10, 18, 0, 5600, 1.0, ist},
		// SNAG 3 — NW-F1 minutes > shift on day +4
		{"NW-PO-1188", nwF1, "released", "Chino press overflow day", 4, 8, 0, 4, 16, 30, 4800, 1.0, ist},
		{"NW-PO-1192", nwF1, "released", "Linen press same day", 4, 9, 0, 4, 15, 30, 3200, 1.0, ist},
		{"NW-PO-1199", nwF1, "released", "Oxford button-card same day", 4, 10, 0, 4, 17, 30, 5600, 1.0, ist},
		{"NW-PO-1210", nwF1, "released", "Autograph dress — later week", 9, 8, 0, 10, 17, 0, 1800, 1.0, ist},
		// Gazipur
		{"GD-PO-7708", gdS1, "sewing", "Hiking short — running hot", -3, 8, 0, 4, 17, 0, 11000, 1.18, ist},
		{"GD-PO-7712", gdS1, "released", "Wide-leg after Decathlon", 6, 8, 0, 10, 17, 0, 4400, 1.0, bst},
		// SNAG 2 — GD-S2 overlap including today
		{"GD-PO-7701", gdS2, "sewing", "Levi's 5-pocket on B line", 0, 8, 0, 3, 17, 0, 9000, 0.72, bst},
		{"GD-PO-7704", gdS2, "loaded", "H&M mom jean dropped on same line", 1, 10, 0, 4, 17, 0, 7200, 1.0, bst},
		{"GD-PO-7701", gdW1, "released", "Levi's stone after sew", 5, 8, 0, 7, 17, 0, 9000, 1.0, bst},
		{"GD-PO-7704", gdW1, "released", "Mom jean vintage recipe", 8, 8, 0, 10, 16, 0, 7200, 1.0, bst},
		{"GD-PO-7708", gdW1, "finishing", "Decathlon enzyme — in house", -1, 8, 0, 1, 16, 0, 4000, 0.9, bst},
	}

	for _, b := range blocks {
		start := at(b.loc, b.d0, b.h0, b.m0)
		end := at(b.loc, b.d1, b.h1, b.m1)
		actual, pieces := initialProgress(b.sop, b.pace, start, end, b.qty)
		id, err := insertBlock(ctx, pool, oids[b.po], b.line, start, end, b.qty, b.sop, b.pace, actual, pieces, b.notes)
		if err != nil {
			return fmt.Errorf("block %s on %s: %w", b.po, b.line, err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO sop_states (block_id, state, actor) VALUES ($1,$2,$3)`, id, b.sop, "seed"); err != nil {
			return err
		}
		if actual > 0 {
			planned := domain.PlannedPct(time.Now(), start, end)
			delay := domain.DelayMinutes(planned, actual, start, end)
			if _, err := pool.Exec(ctx, `
				INSERT INTO progress_ticks (block_id, actual_pct, planned_pct, pieces_done, delay_minutes)
				VALUES ($1,$2,$3,$4,$5)`, id, actual, planned, pieces, delay); err != nil {
				return err
			}
		}
	}
	return nil
}

func initialProgress(sop string, pace float64, start, end time.Time, qty int) (float64, int) {
	now := time.Now()
	planned := domain.PlannedPct(now, start, end)
	if planned == 0 || sop == "released" || sop == "marker" {
		return 0, 0
	}
	actual := planned * pace
	if actual > 100 {
		actual = 100
	}
	if sop == "packed" || sop == "ex_factory" {
		return 100, qty
	}
	if sop == "cut" && actual < 80 {
		actual = 82
	}
	return actual, int(float64(qty) * actual / 100)
}

func insertPlant(ctx context.Context, pool *pgxpool.Pool, code, name, city, region, tz, ss, se string, mins int) (string, error) {
	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO org_plants (code, name, city, region, timezone, shift_start, shift_end, shift_minutes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id::text`, code, name, city, region, tz, ss, se, mins).Scan(&id)
	return id, err
}

func insertUnit(ctx context.Context, pool *pgxpool.Pool, plant, code, name string) (string, error) {
	var id string
	err := pool.QueryRow(ctx, `INSERT INTO units (plant_id, code, name) VALUES ($1,$2,$3) RETURNING id::text`, plant, code, name).Scan(&id)
	return id, err
}

func insertLine(ctx context.Context, pool *pgxpool.Pool, unit, code, name, kind string, ops int, sam float64, mins int) (string, error) {
	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO lines (unit_id, code, name, kind, operators, sam_minutes, shift_minutes)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id::text`, unit, code, name, kind, ops, sam, mins).Scan(&id)
	return id, err
}

func insertOrder(ctx context.Context, pool *pgxpool.Pool, po, style, name, buyer, plant string, qty int, due string, sam float64, color, merch string) (string, error) {
	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO plan_orders (po_code, style_code, style_name, buyer, plant_id, qty, due_date, minutes_per_piece, colorway, merch_owner)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id::text`,
		po, style, name, buyer, plant, qty, due, sam, color, merch).Scan(&id)
	return id, err
}

func insertBlock(ctx context.Context, pool *pgxpool.Pool, order, line string, start, end time.Time, qty int, sop string, pace, actual float64, pieces int, notes string) (string, error) {
	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO gantt_blocks (order_id, line_id, starts_at, ends_at, planned_qty, sop_state, pace, actual_pct, pieces_done, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id::text`,
		order, line, start, end, qty, sop, pace, actual, pieces, notes).Scan(&id)
	return id, err
}
