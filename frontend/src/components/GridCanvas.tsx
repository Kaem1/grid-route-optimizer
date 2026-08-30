import { forwardRef, useEffect, useImperativeHandle, useRef } from "react";
import { type BoardSnapshot, GridRenderer } from "../rendering/GridRenderer";

export interface GridCanvasHandle {
  zoomIn: () => void;
  zoomOut: () => void;
  resetView: () => void;
  fitToGrid: () => void;
}

interface GridCanvasProps {
  snapshot: BoardSnapshot;
  onVertexClick: (id: string) => void;
  onEdgeClick: (key: string, a: string, b: string) => void;
}

export const GridCanvas = forwardRef<GridCanvasHandle, GridCanvasProps>(function GridCanvas(
  { snapshot, onVertexClick, onEdgeClick },
  ref,
) {
  const containerRef = useRef<HTMLDivElement>(null);
  const rendererRef = useRef<GridRenderer | null>(null);
  const dimsRef = useRef({ cols: snapshot.cols, rows: snapshot.rows });

  // Keep the latest callbacks without re-creating the renderer on every render.
  const callbacksRef = useRef({ onVertexClick, onEdgeClick });
  callbacksRef.current = { onVertexClick, onEdgeClick };

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
      onVertexClick: (id) => callbacksRef.current.onVertexClick(id),
      onEdgeClick: (key, a, b) => callbacksRef.current.onEdgeClick(key, a, b),
    });

    renderer.init().then(() => {
      if (cancelled) {
        renderer.destroy();
        return;
      }
      renderer.update(snapshot);
      renderer.fitToGrid();
      rendererRef.current = renderer;
    });

    return () => {
      cancelled = true;
      rendererRef.current = null;
      renderer.destroy();
    };
    // Renderer is created once; subsequent snapshot updates use the effect below.
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

  return <div ref={containerRef} className="grid-canvas" />;
});
