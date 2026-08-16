package http

import (
	"strings"
	"time"

	"garment-ppc/internal/domain"
	"garment-ppc/internal/sim"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	app    *fiber.App
	store  *Store
	ticker *sim.Ticker
	origin string
}

func New(pool *pgxpool.Pool, origin string, ticker *sim.Ticker) *Server {
	s := &Server{
		store:  NewStore(pool),
		ticker: ticker,
		origin: origin,
	}
	app := fiber.New(fiber.Config{
		AppName:      "garment-ppc",
		ErrorHandler: jsonError,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
	})
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: origin,
		AllowHeaders: "Origin, Content-Type, Accept, X-Demo-User, X-Demo-Role",
		AllowMethods: "GET,POST,PATCH,DELETE,OPTIONS",
	}))
	s.app = app
	s.routes()
	return s
}

func (s *Server) App() *fiber.App { return s.app }

func (s *Server) routes() {
	s.app.Get("/health", s.health)
	api := s.app.Group("/api")
	api.Get("/session", s.session)
	api.Get("/users", s.users)
	api.Get("/plants", s.plants)
	api.Get("/units", s.units)
	api.Get("/lines", s.lines)
	api.Get("/orders", s.orders)
	api.Get("/gantt", s.gantt)
	api.Get("/conflicts", s.conflicts)
	api.Get("/dashboard", s.dashboard)
	api.Get("/ticks", s.ticks)
	api.Get("/blocks/:id", s.blockOne)
	api.Get("/blocks/:id/sop", s.sopHistory)
	api.Post("/blocks", s.place)
	api.Patch("/blocks/:id", s.move)
	api.Delete("/blocks/:id", s.remove)
	api.Post("/blocks/:id/sop", s.advance)
	api.Post("/sim/tick", s.forceTick)
}

func jsonError(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
		"hint":  "Check DATABASE_URL, then rerun seed if tables are empty.",
	})
}

func (s *Server) health(c *fiber.Ctx) error {
	ctx := c.UserContext()
	status := "ok"
	db := "up"
	if err := s.store.pool.Ping(ctx); err != nil {
		status = "degraded"
		db = "down"
	}
	seeded, _ := s.store.HasAnyPlant(ctx)
	simState := "idle"
	last := time.Time{}
	if s.ticker != nil {
		simState = s.ticker.Status()
		last = s.ticker.LastTick()
	}
	return c.JSON(fiber.Map{
		"status":  status,
		"db":      db,
		"seeded":  seeded,
		"sim":     simState,
		"last_tick": last,
		"now":     time.Now(),
	})
}

func (s *Server) loadSession(c *fiber.Ctx) (domain.Session, error) {
	slug := c.Get("X-Demo-User")
	if slug == "" {
		slug = c.Query("user")
	}
	if slug == "" {
		slug = "kavya"
	}
	user, err := s.store.UserBySlug(c.UserContext(), slug)
	if err != nil {
		if IsNoRows(err) {
			return domain.Session{}, fiber.NewError(fiber.StatusNotFound, "unknown demo user "+slug)
		}
		return domain.Session{}, err
	}
	return domain.SessionFor(user), nil
}

func (s *Server) session(c *fiber.Ctx) error {
	sess, err := s.loadSession(c)
	if err != nil {
		return err
	}
	users, err := s.store.ListUsers(c.UserContext())
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"session": sess, "users": users})
}

func (s *Server) users(c *fiber.Ctx) error {
	users, err := s.store.ListUsers(c.UserContext())
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"users": users})
}

func (s *Server) plants(c *fiber.Ctx) error {
	sess, err := s.loadSession(c)
	if err != nil {
		return err
	}
	plants, err := s.store.ListPlants(c.UserContext())
	if err != nil {
		return err
	}
	if sess.LockedPlant != nil {
		var filtered []domain.Plant
		for _, p := range plants {
			if p.ID == *sess.LockedPlant {
				filtered = append(filtered, p)
			}
		}
		plants = filtered
	}
	return c.JSON(fiber.Map{"plants": plants})
}

func (s *Server) units(c *fiber.Ctx) error {
	plantID, err := s.scopedPlant(c)
	if err != nil {
		return err
	}
	units, err := s.store.ListUnits(c.UserContext(), plantID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"units": units})
}

func (s *Server) lines(c *fiber.Ctx) error {
	plantID, err := s.scopedPlant(c)
	if err != nil {
		return err
	}
	lines, err := s.store.ListLines(c.UserContext(), plantID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"lines": lines})
}

func (s *Server) orders(c *fiber.Ctx) error {
	sess, err := s.loadSession(c)
	if err != nil {
		return err
	}
	plantID, err := s.scopedPlant(c)
	if err != nil {
		return err
	}
	orders, err := s.store.ListOrders(c.UserContext(), plantID)
	if err != nil {
		return err
	}
	if sess.KPISet == "merch" && sess.FocusBuyer != "" {
		var mine []domain.Order
		var rest []domain.Order
		for _, o := range orders {
			if o.MerchOwner == sess.FocusBuyer {
				mine = append(mine, o)
			} else {
				rest = append(rest, o)
			}
		}
		return c.JSON(fiber.Map{"orders": mine, "other_orders": rest, "focus": sess.FocusBuyer})
	}
	return c.JSON(fiber.Map{"orders": orders})
}

func (s *Server) gantt(c *fiber.Ctx) error {
	sess, err := s.loadSession(c)
	if err != nil {
		return err
	}
	plantID, from, to, err := s.window(c)
	if err != nil {
		return err
	}
	plants, err := s.store.ListPlants(c.UserContext())
	if err != nil {
		return err
	}
	lines, err := s.store.ListLines(c.UserContext(), plantID)
	if err != nil {
		return err
	}
	blocks, err := s.store.ListBlocks(c.UserContext(), plantID, from, to)
	if err != nil {
		return err
	}
	conflicts := domain.DetectConflicts(lines, plants, blocks)
	blocks = domain.ApplyConflicts(blocks, conflicts)
	if sess.LockedPlant != nil {
		var fp []domain.Plant
		for _, p := range plants {
			if p.ID == *sess.LockedPlant {
				fp = append(fp, p)
			}
		}
		plants = fp
	}
	return c.JSON(fiber.Map{
		"now":       time.Now(),
		"from":      from,
		"to":        to,
		"session":   sess,
		"plants":    plants,
		"lines":     lines,
		"blocks":    blocks,
		"conflicts": conflicts,
	})
}

func (s *Server) conflicts(c *fiber.Ctx) error {
	plantID, from, to, err := s.window(c)
	if err != nil {
		return err
	}
	plants, err := s.store.ListPlants(c.UserContext())
	if err != nil {
		return err
	}
	lines, err := s.store.ListLines(c.UserContext(), plantID)
	if err != nil {
		return err
	}
	blocks, err := s.store.ListBlocks(c.UserContext(), plantID, from, to)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"conflicts": domain.DetectConflicts(lines, plants, blocks)})
}

func (s *Server) dashboard(c *fiber.Ctx) error {
	sess, err := s.loadSession(c)
	if err != nil {
		return err
	}
	plantID, from, to, err := s.window(c)
	if err != nil {
		return err
	}
	plants, err := s.store.ListPlants(c.UserContext())
	if err != nil {
		return err
	}
	lines, err := s.store.ListLines(c.UserContext(), plantID)
	if err != nil {
		return err
	}
	blocks, err := s.store.ListBlocks(c.UserContext(), plantID, from, to)
	if err != nil {
		return err
	}
	orders, err := s.store.ListOrders(c.UserContext(), plantID)
	if err != nil {
		return err
	}
	conflicts := domain.DetectConflicts(lines, plants, blocks)
	blocks = domain.ApplyConflicts(blocks, conflicts)

	var delayed, active, packed int
	var actualSum, planSum float64
	for _, b := range blocks {
		if domain.IsActiveSOP(b.SOPState) || (b.ActualPct > 0 && b.ActualPct < 100) {
			active++
			actualSum += b.ActualPct
			planSum += b.PlannedPct
		}
		if b.DelayMinutes >= 30 && b.ActualPct < 100 {
			delayed++
		}
		if b.SOPState == "packed" || b.SOPState == "ex_factory" {
			packed++
		}
	}
	attain := 0.0
	if planSum > 0 {
		attain = (actualSum / planSum) * 100
	}
	overMin := 0
	for _, cf := range conflicts {
		overMin += cf.MinutesOver
	}
	unplaced := 0
	otdRisk := 0
	horizon := time.Now().Add(7 * 24 * time.Hour)
	for _, o := range orders {
		unplaced += o.UnplacedQty
		due, _ := time.Parse("2006-01-02", o.DueDate)
		if !due.After(horizon) && o.UnplacedQty > 0 {
			otdRisk++
		}
	}

	kpis := fiber.Map{
		"collisions":     len(conflicts),
		"lines":          len(lines),
		"active_blocks":  active,
		"delayed_blocks": delayed,
		"plan_attainment": round1(attain),
		"minutes_over":   overMin,
		"packed_blocks":  packed,
		"unplaced_qty":   unplaced,
		"otd_risk":       otdRisk,
		"set":            sess.KPISet,
	}
	return c.JSON(fiber.Map{
		"session":   sess,
		"kpis":      kpis,
		"conflicts": conflicts,
		"now":       time.Now(),
	})
}

func (s *Server) ticks(c *fiber.Ctx) error {
	ticks, err := s.store.RecentTicks(c.UserContext(), c.Query("block_id"), 24)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"ticks": ticks, "last": s.ticker.LastTick()})
}

func (s *Server) blockOne(c *fiber.Ctx) error {
	b, err := s.store.BlockByID(c.UserContext(), c.Params("id"))
	if err != nil {
		if IsNoRows(err) {
			return fiber.NewError(fiber.StatusNotFound, "block not found")
		}
		return err
	}
	hist, err := s.store.SOPHistory(c.UserContext(), b.ID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"block": b, "sop": hist})
}

func (s *Server) sopHistory(c *fiber.Ctx) error {
	hist, err := s.store.SOPHistory(c.UserContext(), c.Params("id"))
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"sop": hist})
}

type placeBody struct {
	OrderID    string  `json:"order_id"`
	LineID     string  `json:"line_id"`
	StartsAt   string  `json:"starts_at"`
	EndsAt     string  `json:"ends_at"`
	PlannedQty int     `json:"planned_qty"`
	SOPState   string  `json:"sop_state"`
	Notes      string  `json:"notes"`
	Pace       float64 `json:"pace"`
}

func (s *Server) place(c *fiber.Ctx) error {
	sess, err := s.loadSession(c)
	if err != nil {
		return err
	}
	if !sess.CanPlace {
		return fiber.NewError(fiber.StatusForbidden, "this role cannot place work on a line")
	}
	var body placeBody
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid place payload")
	}
	line, err := s.store.LineByID(c.UserContext(), body.LineID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "unknown line")
	}
	if !domain.CanMutateLine(sess, line.PlantID) {
		return fiber.NewError(fiber.StatusForbidden, "cannot load a line outside your plant")
	}
	start, end, err := parseSpan(body.StartsAt, body.EndsAt)
	if err != nil {
		return err
	}
	if body.PlannedQty <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "planned_qty must be > 0")
	}
	id, err := s.store.InsertBlock(c.UserContext(), body.OrderID, body.LineID, start, end, body.PlannedQty, body.SOPState, body.Notes, body.Pace)
	if err != nil {
		return err
	}
	b, err := s.store.BlockByID(c.UserContext(), id)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"block": b})
}

type moveBody struct {
	LineID     string `json:"line_id"`
	StartsAt   string `json:"starts_at"`
	EndsAt     string `json:"ends_at"`
	PlannedQty *int   `json:"planned_qty"`
}

func (s *Server) move(c *fiber.Ctx) error {
	sess, err := s.loadSession(c)
	if err != nil {
		return err
	}
	if !sess.CanMove {
		return fiber.NewError(fiber.StatusForbidden, "this role cannot move a pick")
	}
	existing, err := s.store.BlockByID(c.UserContext(), c.Params("id"))
	if err != nil {
		if IsNoRows(err) {
			return fiber.NewError(fiber.StatusNotFound, "block not found")
		}
		return err
	}
	var body moveBody
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid move payload")
	}
	lineID := body.LineID
	if lineID == "" {
		lineID = existing.LineID
	}
	line, err := s.store.LineByID(c.UserContext(), lineID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "unknown line")
	}
	if !domain.CanMutateLine(sess, line.PlantID) && !domain.CanMutateLine(sess, existing.PlantID) {
		return fiber.NewError(fiber.StatusForbidden, "cannot move work outside your plant")
	}
	start, end := existing.StartsAt, existing.EndsAt
	if body.StartsAt != "" && body.EndsAt != "" {
		start, end, err = parseSpan(body.StartsAt, body.EndsAt)
		if err != nil {
			return err
		}
	}
	if err := s.store.UpdateBlock(c.UserContext(), existing.ID, lineID, start, end, body.PlannedQty); err != nil {
		return err
	}
	b, err := s.store.BlockByID(c.UserContext(), existing.ID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"block": b})
}

func (s *Server) remove(c *fiber.Ctx) error {
	sess, err := s.loadSession(c)
	if err != nil {
		return err
	}
	if !sess.CanPlace {
		return fiber.NewError(fiber.StatusForbidden, "this role cannot pull a block")
	}
	if err := s.store.DeleteBlock(c.UserContext(), c.Params("id")); err != nil {
		if IsNoRows(err) {
			return fiber.NewError(fiber.StatusNotFound, "block not found")
		}
		return err
	}
	return c.JSON(fiber.Map{"ok": true})
}

type sopBody struct {
	State  string `json:"state"`
	Action string `json:"action"`
}

func (s *Server) advance(c *fiber.Ctx) error {
	sess, err := s.loadSession(c)
	if err != nil {
		return err
	}
	if !sess.CanAdvance {
		return fiber.NewError(fiber.StatusForbidden, "this role cannot change SOP")
	}
	var body sopBody
	_ = c.BodyParser(&body)
	target := body.State
	if body.Action == "advance" {
		target = ""
	}
	next, err := s.store.AdvanceSOP(c.UserContext(), c.Params("id"), sess.User.Name, target)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	b, err := s.store.BlockByID(c.UserContext(), c.Params("id"))
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"block": b, "sop_state": next})
}

func (s *Server) forceTick(c *fiber.Ctx) error {
	n, err := s.ticker.TickOnce(c.UserContext())
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"updated": n, "last": s.ticker.LastTick()})
}

func (s *Server) scopedPlant(c *fiber.Ctx) (string, error) {
	sess, err := s.loadSession(c)
	if err != nil {
		return "", err
	}
	plantID := c.Query("plant_id")
	if plantID == "" {
		if code := c.Query("plant"); code != "" {
			p, err := s.store.PlantByCode(c.UserContext(), strings.ToUpper(code))
			if err != nil {
				if IsNoRows(err) {
					return "", fiber.NewError(fiber.StatusNotFound, "unknown plant "+code)
				}
				return "", err
			}
			plantID = p.ID
		}
	}
	if sess.LockedPlant != nil {
		if plantID != "" && plantID != *sess.LockedPlant {
			return "", fiber.NewError(fiber.StatusForbidden, "this role is locked to one plant")
		}
		return *sess.LockedPlant, nil
	}
	return plantID, nil
}

func (s *Server) window(c *fiber.Ctx) (plantID string, from, to time.Time, err error) {
	plantID, err = s.scopedPlant(c)
	if err != nil {
		return
	}
	now := time.Now()
	from = startOfDay(now.AddDate(0, 0, -3))
	to = startOfDay(now.AddDate(0, 0, 11))
	if raw := c.Query("from"); raw != "" {
		from, err = time.Parse(time.RFC3339, raw)
		if err != nil {
			from, err = time.Parse("2006-01-02", raw)
			if err != nil {
				err = fiber.NewError(fiber.StatusBadRequest, "from must be YYYY-MM-DD or RFC3339")
				return
			}
		}
	}
	if raw := c.Query("to"); raw != "" {
		to, err = time.Parse(time.RFC3339, raw)
		if err != nil {
			to, err = time.Parse("2006-01-02", raw)
			if err != nil {
				err = fiber.NewError(fiber.StatusBadRequest, "to must be YYYY-MM-DD or RFC3339")
				return
			}
			to = to.Add(24 * time.Hour)
		}
	}
	return
}

func parseSpan(a, b string) (time.Time, time.Time, error) {
	start, err := time.Parse(time.RFC3339, a)
	if err != nil {
		return time.Time{}, time.Time{}, fiber.NewError(fiber.StatusBadRequest, "starts_at must be RFC3339")
	}
	end, err := time.Parse(time.RFC3339, b)
	if err != nil {
		return time.Time{}, time.Time{}, fiber.NewError(fiber.StatusBadRequest, "ends_at must be RFC3339")
	}
	if !end.After(start) {
		return time.Time{}, time.Time{}, fiber.NewError(fiber.StatusBadRequest, "ends_at must be after starts_at")
	}
	return start, end, nil
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func round1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}


