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
  toggleVertex: (id: string) => void;
  toggleEdge: (key: string, a: string, b: string) => void;
  resize: (cols: number, rows: number) => void;
}

const MIN_SIZE = 2;
const MAX_SIZE = 300;

export function useBoardState(initialCols: number, initialRows: number): BoardState {
  const [cols, setCols] = useState(initialCols);
  const [rows, setRows] = useState(initialRows);
  const [activeVertices, setActiveVertices] = useState<Set<string>>(() => new Set());
  const [walls, setWalls] = useState<Set<string>>(() => new Set());

  const toggleVertex = useCallback(
    (id: string) => {
      const wasActive = activeVertices.has(id);

      setActiveVertices((prev) => {
        const next = new Set(prev);
        if (next.has(id)) next.delete(id);
        else next.add(id);
        return next;
      });

      // Deactivating a cell removes any walls attached to it.
      if (wasActive) {
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
      }
    },
    [activeVertices],
  );

  const toggleEdge = useCallback(
    (key: string, a: string, b: string) => {
      if (!activeVertices.has(a) || !activeVertices.has(b)) return;

      setWalls((prev) => {
        const next = new Set(prev);
        if (next.has(key)) next.delete(key);
        else next.add(key);
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
  }, []);

  return { cols, rows, activeVertices, walls, toggleVertex, toggleEdge, resize };
}
