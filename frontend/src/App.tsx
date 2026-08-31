import { useCallback, useMemo, useRef, useState } from "react";
import "./App.css";
import { GridCanvas, type GridCanvasHandle } from "./components/GridCanvas";
import type { EditTool } from "./rendering/GridRenderer";
import { useBoardState } from "./state/useBoardState";

const DEFAULT_COLS = 24;
const DEFAULT_ROWS = 16;

type Mode = "menu" | "edit";

function App() {
  const board = useBoardState(DEFAULT_COLS, DEFAULT_ROWS);
  const canvasRef = useRef<GridCanvasHandle>(null);
  const [backendStatus, setBackendStatus] = useState("Nie sprawdzono");
  const [mode, setMode] = useState<Mode>("menu");
  const [tool, setTool] = useState<EditTool>("select");

  const snapshot = useMemo(
    () => ({
      cols: board.cols,
      rows: board.rows,
      activeVertices: board.activeVertices,
      walls: board.walls,
      startPoints: board.startPoints,
    }),
    [board.cols, board.rows, board.activeVertices, board.walls, board.startPoints],
  );

  const exitEditMode = useCallback(() => {
    setTool("select");
    setMode("menu");
  }, []);

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
          {mode === "menu" ? (
            <>
              <section>
                <h2>Plansza</h2>
                <button className="primary-button" onClick={() => setMode("edit")}>
                  Edytuj planszę
                </button>
              </section>

              <section>
                <h2>Algorytm</h2>
                <select disabled defaultValue="">
                  <option value="" disabled>
                    Wkrótce dostępne
                  </option>
                </select>
              </section>

              <section>
                <h2>Zapisane plansze</h2>
                <select disabled defaultValue="">
                  <option value="" disabled>
                    Wkrótce dostępne
                  </option>
                </select>
              </section>

              <section>
                <h2>Symulacja</h2>
                <button disabled title="Backend jeszcze niedostępny">
                  Uruchom symulację
                </button>
              </section>
            </>
          ) : (
            <>
              <section>
                <h2>Edycja planszy</h2>
                <button className="primary-button" onClick={exitEditMode}>
                  Zapisz
                </button>
                <p className="hint">
                  Lewy przycisk myszy: zaznacz/odznacz kratkę lub ścianę (przytrzymaj i przeciągnij, aby zaznaczać wiele
                  naraz). Prawy przycisk: przesuwanie widoku.
                </p>
              </section>

              <section>
                <h2>Punkty startowe</h2>
                <button
                  className={tool === "start" ? "primary-button" : undefined}
                  onClick={() => setTool(tool === "start" ? "select" : "start")}
                >
                  {tool === "start" ? "Zakończ dodawanie" : "Dodaj punkt startowy"}
                </button>
                <p className="hint">Kliknij zaznaczoną kratkę, aby dodać lub usunąć punkt startowy.</p>
              </section>

              <section>
                <h2>Rozmiar kraty</h2>
                <label>
                  Kolumny
                  <input
                    type="number"
                    min={2}
                    max={300}
                    value={board.cols}
                    onChange={(e) => board.resize(Number(e.target.value), board.rows)}
                  />
                </label>
                <label>
                  Wiersze
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
                <h2>Narzędzia (wkrótce)</h2>
                <div className="tool-buttons">
                  <button disabled title="Wkrótce dostępne">
                    Prostokąt
                  </button>
                  <button disabled title="Wkrótce dostępne">
                    Elipsa
                  </button>
                </div>
              </section>
            </>
          )}

          <section>
            <h2>Widok</h2>
            <div className="view-buttons">
              <button onClick={() => canvasRef.current?.zoomIn()}>+</button>
              <button onClick={() => canvasRef.current?.zoomOut()}>−</button>
              <button onClick={() => canvasRef.current?.fitToGrid()}>Dopasuj</button>
              <button onClick={() => canvasRef.current?.resetView()}>Reset</button>
            </div>
          </section>

          <section>
            <h2>Legenda</h2>
            <ul className="legend">
              <li>
                <span className="swatch swatch-active" /> Aktywna kratka
              </li>
              <li>
                <span className="swatch swatch-wall" /> Ściana
              </li>
              <li>
                <span className="swatch swatch-start" /> Punkt startowy
              </li>
            </ul>
          </section>
        </aside>

        <main className="grid-area">
          <GridCanvas
            ref={canvasRef}
            snapshot={snapshot}
            editable={mode === "edit"}
            tool={tool}
            onVertexPaint={board.setVertexActive}
            onEdgePaint={board.setWallActive}
            onStartPointPaint={board.setStartPointActive}
          />
        </main>
      </div>

      <footer className="statusbar">
        <span>Aktywne kratki: {board.activeVertices.size}</span>
        <span>Ściany: {board.walls.size}</span>
        <span>Backend: {backendStatus}</span>
      </footer>
    </div>
  );
}

export default App;