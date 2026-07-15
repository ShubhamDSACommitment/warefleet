"""Prioritized planning MAPF (baseline).

Prioritized planning is the simplest usable MAPF method and the Week-5 target:
  1. order the agents by priority,
  2. plan each agent's shortest path in turn with space-time A*,
  3. treat already-planned agents' paths as moving obstacles (reservation table).

It is fast and incomplete (can fail on tightly-coupled instances) — which is
exactly why it makes a good BASELINE to compare CBS/PIBT against in the study.

Prototype and unit-test the algorithm here (pure Python, fast to iterate),
then port the validated logic into the Go coordinator (fleet_manager).

References to read first:
  * Silver, "Cooperative Pathfinding" (2005) — space-time A*, reservation table
  * Sharon et al., "Conflict-Based Search" (2015) — for the CBS stretch goal
  * Okumura et al., "PIBT" — for the scalable stretch goal
"""
from __future__ import annotations

from dataclasses import dataclass, field

Cell = tuple[int, int]          # (row, col) on the warehouse grid
TimedCell = tuple[int, int, int]  # (row, col, timestep)


@dataclass
class Agent:
    agent_id: str
    start: Cell
    goal: Cell
    priority: int = 0


@dataclass
class Reservation:
    """Space-time reservation table of cells occupied by already-planned agents."""
    occupied: set[TimedCell] = field(default_factory=set)
    # edge reservations prevent head-on swaps: frozenset({a, b}) at time t
    edges: set[tuple[frozenset, int]] = field(default_factory=set)

    def reserve(self, path: list[Cell]) -> None:
        for t, cell in enumerate(path):
            self.occupied.add((cell[0], cell[1], t))
        for t in range(len(path) - 1):
            self.edges.add((frozenset({path[t], path[t + 1]}), t))

    def is_free(self, cell: Cell, t: int) -> bool:
        return (cell[0], cell[1], t) not in self.occupied

    def edge_free(self, a: Cell, b: Cell, t: int) -> bool:
        return (frozenset({a, b}), t) not in self.edges


def space_time_astar(grid, start: Cell, goal: Cell, res: Reservation,
                     max_t: int = 512) -> list[Cell] | None:
    """Shortest path from start to goal that respects the reservation table.

    TODO(week5): implement space-time A*:
      * state = (cell, t); neighbours = 4-connected moves + WAIT
      * skip a neighbour if not res.is_free(cell, t+1) or not edge_free
      * heuristic = manhattan(cell, goal); goal reached when at goal and can wait
    Returns the path as a list of cells indexed by timestep, or None if no path.
    """
    raise NotImplementedError('space_time_astar: implement in Week 5')


def plan_prioritized(grid, agents: list[Agent]) -> dict[str, list[Cell]]:
    """Plan all agents by descending priority into a shared reservation table."""
    res = Reservation()
    plans: dict[str, list[Cell]] = {}
    for agent in sorted(agents, key=lambda a: -a.priority):
        path = space_time_astar(grid, agent.start, agent.goal, res)
        if path is None:
            raise RuntimeError(f'no conflict-free path for {agent.agent_id}')
        res.reserve(path)
        plans[agent.agent_id] = path
    return plans
