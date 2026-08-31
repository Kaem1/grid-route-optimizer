import { forwardRef, useEffect, useImperativeHandle, useRef } from "react";
import { type BoardSnapshot, type EditTool, GridRenderer, type SimulationVisual } from "../rendering/GridRenderer";

export interface GridCanvasHandle {
  zoomIn: () => void;
  zoomOut: () => void;
  resetView: () => void;
  fitToGrid: () => void;
}

interface GridCanvasProps {
  snapshot: BoardSnapshot;
  editable: boolean;
  tool: EditTool;
  simulation: SimulationVisual | null;
  onVertexPaint: (id: string, active: boolean) => void;
  onEdgePaint: (key: string, a: string, b: string, active: boolean) => void;
  onStartPointPaint: (id: string, active: boolean) => void;
}

export const GridCanvas = forwardRef<GridCanvasHandle, GridCanvasProps>(function GridCanvas(
  { snapshot, editable, tool, simulation, onVertexPaint, onEdgePaint, onStartPointPaint },
  ref,
) {
  const containerRef = useRef<HTMLDivElement>(null);
  const rendererRef = useRef<GridRenderer | null>(null);
  const dimsRef = useRef({ cols: snapshot.cols, rows: snapshot.rows });

  // Keep the latest callbacks without re-creating the renderer on every render.
  const callbacksRef = useRef({ onVertexPaint, onEdgePaint, onStartPointPaint });
  callbacksRef.current = { onVertexPaint, onEdgePaint, onStartPointPaint };

  useImperativeHandle(ref, () => ({
    zoomIn: () => rendererRef.current?.zoomIn(),
    zoomOut: () => rendererRef.current?.zoomOut(),
    resetView: () => rendererRef.current?.resetView(),
    fitToGrid: () => rendererRef.current?.fitToGrid(),
  }));

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    let cancelled = false;
    const renderer = new GridRenderer(container, {
      onVertexPaint: (id, active) => callbacksRef.current.onVertexPaint(id, active),
      onEdgePaint: (key, a, b, active) => callbacksRef.current.onEdgePaint(key, a, b, active),
      onStartPointPaint: (id, active) => callbacksRef.current.onStartPointPaint(id, active),
    });

    renderer.init().then(() => {
      if (cancelled) {
        renderer.destroy();
        return;
      }
      renderer.update(snapshot);
      renderer.setEditable(editable);
      renderer.setTool(tool);
      renderer.updateSimulation(simulation);
      renderer.fitToGrid();
      rendererRef.current = renderer;
    });

    return () => {
      cancelled = true;
      rendererRef.current = null;
      renderer.destroy();
    };
    // Renderer is created once; subsequent snapshot/editable/tool updates use the effects below.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    const renderer = rendererRef.current;
    if (!renderer) return;

    renderer.update(snapshot);

    if (dimsRef.current.cols !== snapshot.cols || dimsRef.current.rows !== snapshot.rows) {
      dimsRef.current = { cols: snapshot.cols, rows: snapshot.rows };
      renderer.fitToGrid();
    }
  }, [snapshot]);

  useEffect(() => {
    rendererRef.current?.setEditable(editable);
  }, [editable]);

  useEffect(() => {
    rendererRef.current?.setTool(tool);
  }, [tool]);

  useEffect(() => {
    rendererRef.current?.updateSimulation(simulation);
  }, [simulation]);

  return <div ref={containerRef} className="grid-canvas" />;
});
