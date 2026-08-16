package sim

import (
	"context"
	"log"
	"sync"
	"time"

	"garment-ppc/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Ticker struct {
	pool     *pgxpool.Pool
	every    time.Duration
	mu       sync.Mutex
	last     time.Time
	running  bool
	stop     chan struct{}
}

func New(pool *pgxpool.Pool, every time.Duration) *Ticker {
	if every < time.Second {
		every = 4 * time.Second
	}
	return &Ticker{pool: pool, every: every, stop: make(chan struct{})}
}

func (t *Ticker) Start() {
	t.mu.Lock()
	if t.running {
		t.mu.Unlock()
		return
	}
	t.running = true
	t.mu.Unlock()
	go t.loop()
}

func (t *Ticker) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.running {
		return
	}
	close(t.stop)
	t.running = false
}

func (t *Ticker) Status() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.running {
		return "ticking"
	}
	return "stopped"
}

func (t *Ticker) LastTick() time.Time {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.last
}

func (t *Ticker) loop() {
	ticker := time.NewTicker(t.every)
	defer ticker.Stop()
	for {
		select {
		case <-t.stop:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			if _, err := t.TickOnce(ctx); err != nil {
				log.Printf("sim tick: %v", err)
			}
			cancel()
		}
	}
}

func (t *Ticker) TickOnce(ctx context.Context) (int, error) {
	now := time.Now()
	rows, err := t.pool.Query(ctx, `
		SELECT id::text, starts_at, ends_at, planned_qty, sop_state, pace::float8, actual_pct::float8, pieces_done
		FROM gantt_blocks
		WHERE actual_pct < 100
		  AND sop_state IN ('loaded','sewing','mid_qc','finishing')
		  AND starts_at <= $1`, now)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type row struct {
		id       string
		start    time.Time
		end      time.Time
		qty      int
		sop      string
		pace     float64
		actual   float64
		pieces   int
	}
	var batch []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.start, &r.end, &r.qty, &r.sop, &r.pace, &r.actual, &r.pieces); err != nil {
			return 0, err
		}
		batch = append(batch, r)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	updated := 0
	for _, r := range batch {
		planned := domain.PlannedPct(now, r.start, r.end)
		dur := r.end.Sub(r.start)
		if dur <= 0 {
			continue
		}
		step := r.pace * (t.every.Seconds() / dur.Seconds()) * 100
		if step < 0.05 {
			step = 0.05
		}
		if step > 4 {
			step = 4
		}
		next := r.actual + step
		if next > 100 {
			next = 100
		}
		// Do not run far ahead of wall-clock plan.
		if next > planned+6 {
			next = planned + 6
			if next > 100 {
				next = 100
			}
		}
		pieces := int(float64(r.qty) * next / 100)
		delay := domain.DelayMinutes(planned, next, r.start, r.end)
		if _, err := t.pool.Exec(ctx, `
			UPDATE gantt_blocks SET actual_pct = $2, pieces_done = $3 WHERE id = $1`,
			r.id, next, pieces); err != nil {
			return updated, err
		}
		if _, err := t.pool.Exec(ctx, `
			INSERT INTO progress_ticks (block_id, actual_pct, planned_pct, pieces_done, delay_minutes)
			VALUES ($1,$2,$3,$4,$5)`, r.id, next, planned, pieces, delay); err != nil {
			return updated, err
		}
		updated++
	}

	t.mu.Lock()
	t.last = now
	t.mu.Unlock()
	return updated, nil
}
