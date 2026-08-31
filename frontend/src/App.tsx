import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import "./App.css";
import { getAlgorithms, runSimulation } from "./api/simulationApi";
import { GridCanvas, type GridCanvasHandle } from "./components/GridCanvas";
import type { EditTool } from "./rendering/GridRenderer";
import { useBoardState } from "./state/useBoardState";
import { useSimulation } from "./state/useSimulation";

const DEFAULT_COLS = 24;
const DEFAULT_ROWS = 16;

type Mode = "menu" | "edit" | "simulation";

function App() {
  const board = useBoardState(DEFAULT_COLS, DEFAULT_ROWS);
  const simulation = useSimulation();
  const canvasRef = useRef<GridCanvasHandle>(null);
  const [backendStatus, setBackendStatus] = useState("Nie sprawdzono");
  const [mode, setMode] = useState<Mode>("menu");
  const [tool, setTool] = useState<EditTool>("select");
  const [algorithms, setAlgorithms] = useState<string[]>([]);
  const [selectedAlgorithm, setSelectedAlgorithm] = useState("");
  const [isStartingSimulation, setIsStartingSimulation] = useState(false);
  const [simulationError, setSimulationError] = useState<string | null>(null);

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

  useEffect(() => {
    getAlgorithms()
      .then((response) => {
        setAlgorithms(response.algorithms);
        setSelectedAlgorithm((current) => current || response.algorithms[0] || "");
      })
      .catch((error) => console.error("Failed to load algorithms", error));
  }, []);

  const exitEditMode = useCallback(() => {
    setTool("select");
    setMode("menu");
  }, []);

  const canRunSimulation =
    board.activeVertices.size > 0 && board.startPoints.size > 0 && selectedAlgorithm !== "" && !isStartingSimulation;

  const handleRunSimulation = useCallback(async () => {
    if (!canRunSimulation) return;

    setIsStartingSimulation(true);
    setSimulationError(null);
    try {
      const response = await runSimulation(
        {
          cols: board.cols,
          rows: board.rows,
          activeVertices: Array.from(board.activeVertices),
          walls: Array.from(board.walls),
        },
        Array.from(board.startPoints),
        selectedAlgorithm,
      );
      simulation.start(response);
      setMode("simulation");
    } catch (error) {
      setSimulationError(error instanceof Error ? error.message : "Nie udało się uruchomić symulacji");
    } finally {
      setIsStartingSimulation(false);
    }
  }, [canRunSimulation, board.cols, board.rows, board.activeVertices, board.walls, board.startPoints, selectedAlgorithm, simulation]);

  const handleEndSimulation = useCallback(() => {
    simulation.stop();
    setMode("menu");
  }, [simulation]);

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
          {mode === "menu" && (
            <>
              <section>
                <h2>Plansza</h2>
                <button className="primary-button" onClick={() => setMode("edit")}>
                  Edytuj planszę
                </button>
              </section>

              <section>
                <h2>Algorytm</h2>
                <select
                  value={selectedAlgorithm}
                  disabled={algorithms.length === 0}
                  onChange={(e) => setSelectedAlgorithm(e.target.value)}
                >
                  {algorithms.length === 0 ? (
                    <option value="">Ładowanie…</option>
                  ) : (
                    algorithms.map((name) => (
                      <option key={name} value={name}>
                        {name}
                      </option>
                    ))
                  )}
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
                <button onClick={handleRunSimulation} disabled={!canRunSimulation}>
                  {isStartingSimulation ? "Obliczanie…" : "Uruchom symulację"}
                </button>
                {!canRunSimulation && !isStartingSimulation && (
                  <p className="hint">Zaznacz kratki i dodaj co najmniej jeden punkt startowy w edytorze.</p>
                )}
                {simulationError && <p className="hint hint-error">{simulationError}</p>}
              </section>
            </>
          )}

          {mode === "edit" && (
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

          {mode === "simulation" && (
            <>
              <section>
                <h2>Symulacja</h2>
                <p className="hint">Algorytm: {simulation.algorithm}</p>
                <p className="hint">
                  Krok {simulation.step} / {simulation.maxStep}
                </p>
                <div className="tool-buttons">
                  <button onClick={simulation.stepBackward} disabled={simulation.step === 0}>
                    ◀ Wstecz
                  </button>
                  <button onClick={simulation.stepForward} disabled={simulation.step >= simulation.maxStep}>
                    Dalej ▶
                  </button>
                </div>
                <button
                  className={simulation.isPlaying ? "primary-button" : undefined}
                  onClick={simulation.togglePlay}
                  disabled={simulation.step >= simulation.maxStep && !simulation.isPlaying}
                >
                  {simulation.isPlaying ? "Pauza" : "Odtwarzaj"}
                </button>
                <label className="speed-label">
                  Szybkość ({simulation.speed} {simulation.speed === 1 ? "krok/s" : "kroki/s"})
                  <input
                    type="range"
                    min={1}
                    max={10}
                    value={simulation.speed}
                    onChange={(e) => simulation.setSpeed(Number(e.target.value))}
                  />
                </label>
                <button onClick={handleEndSimulation}>Zakończ symulację</button>
              </section>

              {simulation.metrics && (
                <section>
                  <h2>Statystyki</h2>
                  <ul className="legend">
                    <li>Trasy: {simulation.metrics.numberOfRoutes}</li>
                    <li>Łączna długość: {simulation.metrics.totalLength}</li>
                    <li>Pokryte kratki: {simulation.metrics.coveredVertices}</li>
                    <li>Pokrycie: {(simulation.metrics.coverageRatio * 100).toFixed(0)}%</li>
                  </ul>
                </section>
              )}
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
            simulation={mode === "simulation" ? simulation.snapshot : null}
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