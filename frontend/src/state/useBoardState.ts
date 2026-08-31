import { useCallback, useState } from "react";
import { edgeTouchesVertex } from "../graph/grid";

// UI-facing board state: which cells belong to the board and where walls are.
// This is interaction state, not the domain graph model owned by the backend
// (see project instructions, section 66) — it only tracks what the user has
// selected so it can later be sent to the backend as a subgraph definition.
export interface BoardState {
  cols: number;
  rows: number;
  activeVertices: ReadonlySet<string>;
  walls: ReadonlySet<string>;
  startPoints: ReadonlySet<string>;
  setVertexActive: (id: string, active: boolean) => void;
  setWallActive: (key: string, a: string, b: string, active: boolean) => void;
  setStartPointActive: (id: string, active: boolean) => void;
  resize: (cols: number, rows: number) => void;
}

const MIN_SIZE = 2;
const MAX_SIZE = 300;

export function useBoardState(initialCols: number, initialRows: number): BoardState {
  const [cols, setCols] = useState(initialCols);
  const [rows, setRows] = useState(initialRows);
  const [activeVertices, setActiveVertices] = useState<Set<string>>(() => new Set());
  const [walls, setWalls] = useState<Set<string>>(() => new Set());
  const [startPoints, setStartPoints] = useState<Set<string>>(() => new Set());

  const setVertexActive = useCallback((id: string, active: boolean) => {
    setActiveVertices((prev) => {
      if (prev.has(id) === active) return prev;
      const next = new Set(prev);
      if (active) next.add(id);
      else next.delete(id);
      return next;
    });

    // Deactivating a cell removes any walls and start point attached to it.
    if (!active) {
      setWalls((prev) => {
        let changed = false;
        const next = new Set(prev);
        for (const key of prev) {
          if (edgeTouchesVertex(key, id)) {
            next.delete(key);
            changed = true;
          }
        }
        return changed ? next : prev;
      });

      setStartPoints((prev) => {
        if (!prev.has(id)) return prev;
        const next = new Set(prev);
        next.delete(id);
        return next;
      });
    }
  }, []);

  const setWallActive = useCallback(
    (key: string, a: string, b: string, active: boolean) => {
      if (!activeVertices.has(a) || !activeVertices.has(b)) return;

      setWalls((prev) => {
        if (prev.has(key) === active) return prev;
        const next = new Set(prev);
        if (active) next.add(key);
        else next.delete(key);
        return next;
      });
    },
    [activeVertices],
  );

  const setStartPointActive = useCallback(
    (id: string, active: boolean) => {
      if (active && !activeVertices.has(id)) return;

      setStartPoints((prev) => {
        if (prev.has(id) === active) return prev;
        const next = new Set(prev);
        if (active) next.add(id);
        else next.delete(id);
        return next;
      });
    },
    [activeVertices],
  );

  const resize = useCallback((newCols: number, newRows: number) => {
    const clampedCols = Math.min(Math.max(Math.round(newCols) || MIN_SIZE, MIN_SIZE), MAX_SIZE);
    const clampedRows = Math.min(Math.max(Math.round(newRows) || MIN_SIZE, MIN_SIZE), MAX_SIZE);
    setCols(clampedCols);
    setRows(clampedRows);
    setActiveVertices(new Set());
    setWalls(new Set());
    setStartPoints(new Set());
  }, []);

  return {
    cols,
    rows,
    activeVertices,
    walls,
    startPoints,
    setVertexActive,
    setWallActive,
    setStartPointActive,
    resize,
  };
}
