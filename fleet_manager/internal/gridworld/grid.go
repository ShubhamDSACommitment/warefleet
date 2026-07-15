// Package gridworld is the idealized tier of the two-simulator design: the
// warehouse as a 2D grid, plus space-time A* with a reservation table — the
// building blocks of prioritized MAPF (Silver 2005, "Cooperative Pathfinding").
//
// The SAME warehouse floor plan as worlds/warehouse.sdf (layout contract),
// discretised at 0.5 m/cell in the WORLD frame. Tasks arrive in the map frame
// (what Nav2 and the scenarios use); ToCell converts map->world->cell.
package gridworld

import "math"

const (
	CellSize = 0.5 // metres per grid cell
	// map frame = world frame + this offset (the SLAM run that built the
	// Nav2 map started at world (-2.0,-0.5), which became map (0,0)).
	MapDX = 2.0
	MapDY = 0.5
)

// Cell is a grid coordinate (col x, row y).
type Cell struct{ X, Y int }

// Grid is a static occupancy grid of the warehouse in the world frame.
type Grid struct {
	W, H    int
	originX float64 // world coords of cell (0,0)'s corner
	originY float64
	blocked []bool
}

func (g *Grid) idx(c Cell) int       { return c.Y*g.W + c.X }
func (g *Grid) InBounds(c Cell) bool { return c.X >= 0 && c.X < g.W && c.Y >= 0 && c.Y < g.H }
func (g *Grid) Blocked(c Cell) bool  { return !g.InBounds(c) || g.blocked[g.idx(c)] }

// block marks all cells intersecting the world-frame axis-aligned rectangle.
func (g *Grid) block(x0, y0, x1, y1 float64) {
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			cx0 := g.originX + float64(x)*CellSize
			cy0 := g.originY + float64(y)*CellSize
			if cx0 < x1 && cx0+CellSize > x0 && cy0 < y1 && cy0+CellSize > y0 {
				g.blocked[y*g.W+x] = true
			}
		}
	}
}

// ToCell converts MAP-frame coordinates (scenario/Nav2 convention) to a cell.
func (g *Grid) ToCell(mapX, mapY float64) Cell {
	wx, wy := mapX-MapDX, mapY-MapDY
	return Cell{
		X: int(math.Floor((wx - g.originX) / CellSize)),
		Y: int(math.Floor((wy - g.originY) / CellSize)),
	}
}

// ToMap returns the map-frame centre of a cell (for reporting).
func (g *Grid) ToMap(c Cell) (float64, float64) {
	wx := g.originX + (float64(c.X)+0.5)*CellSize
	wy := g.originY + (float64(c.Y)+0.5)*CellSize
	return wx + MapDX, wy + MapDY
}

// NearestFree returns c, or the nearest unblocked cell (spiral search) — task
// stations sit close to racks and can land on a blocked cell after rounding.
func (g *Grid) NearestFree(c Cell) Cell {
	return g.NearestFreeExcluding(c, nil)
}

// NearestFreeExcluding is NearestFree that also treats `avoid` cells (e.g.
// parked robots) as blocked.
func (g *Grid) NearestFreeExcluding(c Cell, avoid map[Cell]bool) Cell {
	ok := func(n Cell) bool { return g.InBounds(n) && !g.Blocked(n) && !avoid[n] }
	if ok(c) {
		return c
	}
	for r := 1; r < max(g.W, g.H); r++ {
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				if abs(dx) != r && abs(dy) != r {
					continue
				}
				if n := (Cell{c.X + dx, c.Y + dy}); ok(n) {
					return n
				}
			}
		}
	}
	return c
}

// NewWarehouse builds the grid matching worlds/warehouse.sdf's layout contract:
// 12m x 8m room, two 4m x 0.8m racks, two staging pallets, one box pile.
func NewWarehouse() *Grid {
	g := &Grid{
		W:       24, // 12 m / 0.5
		H:       16, // 8 m / 0.5
		originX: -6, originY: -4,
	}
	g.blocked = make([]bool, g.W*g.H)

	// outer walls (0.2 thick, centred on the boundary — block the rim cells)
	g.block(-6.0, -4.0, 6.0, -3.9) // south
	g.block(-6.0, 3.9, 6.0, 4.0)   // north
	g.block(-6.0, -4.0, -5.9, 4.0) // west
	g.block(5.9, -4.0, 6.0, 4.0)   // east
	// rack A (world x[-4,0], y[1.1,1.9]) and rack B (x[0,4], y[-1.9,-1.1])
	g.block(-4.0, 1.1, 0.0, 1.9)
	g.block(0.0, -1.9, 4.0, -1.1)
	// staging pallets (SW at (-4.8,-3) rotated 90°, NE at (4.8,3)) + box pile
	g.block(-5.2, -3.55, -4.4, -2.45)
	g.block(4.25, 2.6, 5.35, 3.4)
	g.block(-0.1, 3.1, 1.1, 3.8)
	return g
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
