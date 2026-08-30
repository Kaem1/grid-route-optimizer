import { useCallback, useMemo, useRef, useState } from "react";
import "./App.css";
import { GridCanvas, type GridCanvasHandle } from "./components/GridCanvas";
import { useBoardState } from "./state/useBoardState";

const DEFAULT_COLS = 24;
const DEFAULT_ROWS = 16;

function App() {
  const board = useBoardState(DEFAULT_COLS, DEFAULT_ROWS);
  const canvasRef = useRef<GridCanvasHandle>(null);
  const [backendStatus, setBackendStatus] = useState("Nie sprawdzono");

  const snapshot = useMemo(
    () => ({ cols: board.cols, rows: board.rows, activeVertices: board.activeVertices, walls: board.walls }),
    [board.cols, board.rows, board.activeVertices, board.walls],
  );

  const checkBackend = useCallback(async () => {
    try {
      const response = await fetch("http://localhost:8080/api/health");

      if (!response.ok) {
        throw new Error("Backend zwrócił błąd");
      }

      const data = await response.json();
      setBackendStatus(data.status);
    } catch (error) {
      console.error(error);
      setBackendStatus("Błąd połączenia");
    }
  }, []);

  return (
    <div className="app">
      <header className="topbar">
        <h1>Grid Route Optimizer</h1>
        <button onClick={checkBackend}>Sprawdź backend</button>
      </header>

      <div className="workspace">
        <aside className="side-panel">
          <section>
            <h2>Grid</h2>
            <label>
              Columns
              <input
                type="number"
                min={2}
                max={300}
                value={board.cols}
                onChange={(e) => board.resize(Number(e.target.value), board.rows)}
              />
            </label>
            <label>
              Rows
              <input
                type="number"
                min={2}
                max={300}
                value={board.rows}
                onChange={(e) => board.resize(board.cols, Number(e.target.value))}
              />
            </label>
          </section>

          <section>
            <h2>View</h2>
            <div className="view-buttons">
              <button onClick={() => canvasRef.current?.zoomIn()}>+</button>
              <button onClick={() => canvasRef.current?.zoomOut()}>−</button>
              <button onClick={() => canvasRef.current?.fitToGrid()}>Fit</button>
              <button onClick={() => canvasRef.current?.resetView()}>Reset</button>
            </div>
          </section>

          <section>
            <h2>Legend</h2>
            <ul className="legend">
              <li>
                <span className="swatch swatch-active" /> Active cell
              </li>
              <li>
                <span className="swatch swatch-wall" /> Wall
              </li>
            </ul>
          </section>
        </aside>

        <main className="grid-area">
          <GridCanvas ref={canvasRef} snapshot={snapshot} onVertexClick={board.toggleVertex} onEdgeClick={board.toggleEdge} />
        </main>
      </div>

      <footer className="statusbar">
        <span>Active cells: {board.activeVertices.size}</span>
        <span>Walls: {board.walls.size}</span>
        <span>Backend: {backendStatus}</span>
      </footer>
    </div>
  );
}

export default App;