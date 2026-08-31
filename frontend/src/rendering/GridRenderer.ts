import { Application, Container, type FederatedPointerEvent, Graphics } from "pixi.js";
import { edgeEndpoints, edgeKey } from "../graph/grid";

// Visual size of a single cell in world units. Purely a rendering concern.
export const CELL_SIZE = 48;

const EDGE_BAND = 10;
const MIN_SCALE = 0.2;
const MAX_SCALE = 4;
const WALL_THICKNESS = CELL_SIZE * 0.34;

const COLOR_GRID_LINE = 0x333333;
const COLOR_CELL_ACTIVE = 0xced3da;
const COLOR_WALL = 0x585d66;
const COLOR_HOVER = 0xffffff;
const COLOR_START = 0xffd23f;
const START_RADIUS = CELL_SIZE * 0.3;

// How strongly a cell's visited-overlay color intensifies with repeat visits
// (by the same or different agents), capped so it never fully hides the grid.
const VISIT_BASE_ALPHA = 0.28;
const VISIT_ALPHA_STEP = 0.14;
const VISIT_MAX_ALPHA = 0.85;

const TRAIL_WIDTH = CELL_SIZE * 0.06;
const TRAIL_ALPHA = 0.45;

const AGENT_RADIUS = CELL_SIZE * 0.2;
const AGENT_STROKE_COLOR = 0x1a1a1a;

export type EditTool = "select" | "start";

export interface BoardSnapshot {
  cols: number;
  rows: number;
  activeVertices: ReadonlySet<string>;
  walls: ReadonlySet<string>;
  startPoints: ReadonlySet<string>;
}

// Playback visual for one simulation step: which cells have been visited (and
// how intensely), the trail segments walked so far, and each agent's current
// position. GridRenderer only draws this; the step engine lives in
// frontend/src/state/useSimulation.ts.
export interface SimulationVisual {
  visited: ReadonlyMap<string, { color: number; count: number }>;
  trail: ReadonlyArray<{ from: string; to: string; color: number }>;
  agents: ReadonlyArray<{ id: number; vertex: string; color: number }>;
}

export interface GridRendererCallbacks {
  onVertexPaint: (id: string, active: boolean) => void;
  onEdgePaint: (key: string, a: string, b: string, active: boolean) => void;
  onStartPointPaint: (id: string, active: boolean) => void;
}

type HitResult =
  | { kind: "vertex"; id: string }
  | { kind: "edge"; key: string; a: string; b: string }
  | { kind: "none" };

function hitEquals(a: HitResult, b: HitResult): boolean {
  if (a.kind !== b.kind) return false;
  if (a.kind === "vertex" && b.kind === "vertex") return a.id === b.id;
  if (a.kind === "edge" && b.kind === "edge") return a.key === b.key;
  return true;
}

function parseVertexId(id: string): [number, number] {
  const [x, y] = id.split(",").map(Number);
  return [x, y];
}

// Rectangle covering the shared border between two adjacent vertices.
function edgeRect(aId: string, bId: string, thickness: number): { x: number; y: number; w: number; h: number } {
  const [ax, ay] = parseVertexId(aId);
  const [bx, by] = parseVertexId(bId);

  if (ax !== bx) {
    const borderX = Math.max(ax, bx) * CELL_SIZE;
    const y = Math.min(ay, by) * CELL_SIZE;
    return { x: borderX - thickness / 2, y, w: thickness, h: CELL_SIZE };
  }

  const borderY = Math.max(ay, by) * CELL_SIZE;
  const x = Math.min(ax, bx) * CELL_SIZE;
  return { x, y: borderY - thickness / 2, w: CELL_SIZE, h: thickness };
}

// Renders the grid board with PixiJS and handles pan/zoom/hover/click interaction.
// This class owns no domain logic (see project instructions, section 65): it only
// reports raw vertex/edge hits and displays whatever BoardSnapshot it is given.
export class GridRenderer {
  private readonly container: HTMLElement;
  private readonly callbacks: GridRendererCallbacks;

  private readonly app = new Application();
  private readonly world = new Container();
  private readonly gridLayer = new Graphics();
  private readonly cellsLayer = new Graphics();
  private readonly visitLayer = new Graphics();
  private readonly wallsLayer = new Graphics();
  private readonly trailLayer = new Graphics();
  private readonly startLayer = new Graphics();
  private readonly agentLayer = new Graphics();
  private readonly hoverLayer = new Graphics();

  private snapshot: BoardSnapshot = { cols: 0, rows: 0, activeVertices: new Set(), walls: new Set(), startPoints: new Set() };
  private simulation: SimulationVisual | null = null;

  private scale = 1;
  private offsetX = 0;
  private offsetY = 0;

  // Whether the board can currently be edited (left-button paint). Panning/zoom always work.
  private editable = true;

  // Which left-button action is currently selected: normal select/paint, or start-point placement.
  private tool: EditTool = "select";

  // Right-button drag pans the view.
  private isPanning = false;
  private panStartScreen = { x: 0, y: 0 };
  private panStartOffset = { x: 0, y: 0 };

  // Left-button drag keeps applying the same action (add/remove) to whatever
  // kind of target (vertex or edge) was under the pointer on mouse down.
  // "eraseStart" is a special stroke: the first click on a cell that has a
  // start-point marker only clears the marker instead of touching the cell.
  private paintMode: "vertex" | "edge" | "start" | "eraseStart" | null = null;
  private paintValue = false;
  private paintedIds = new Set<string>();

  private hovered: HitResult = { kind: "none" };
  private initialized = false;

  constructor(container: HTMLElement, callbacks: GridRendererCallbacks) {
    this.container = container;
    this.callbacks = callbacks;
  }

  async init(): Promise<void> {
    await this.app.init({
      background: 0x000000,
      resizeTo: this.container,
      antialias: true,
      resolution: window.devicePixelRatio || 1,
      autoDensity: true,
    });

    this.container.appendChild(this.app.canvas);

    this.world.addChild(
      this.gridLayer,
      this.cellsLayer,
      this.visitLayer,
      this.wallsLayer,
      this.trailLayer,
      this.startLayer,
      this.agentLayer,
      this.hoverLayer,
    );
    this.app.stage.addChild(this.world);

    this.app.stage.eventMode = "static";
    this.app.stage.hitArea = this.app.screen;

    this.app.stage.on("pointerdown", this.handlePointerDown);
    this.app.stage.on("pointermove", this.handlePointerMove);
    this.app.stage.on("pointerup", this.handlePointerUp);
    this.app.stage.on("pointerupoutside", this.handlePointerUp);
    this.app.canvas.addEventListener("wheel", this.handleWheel, { passive: false });
    this.app.canvas.addEventListener("contextmenu", this.handleContextMenu);

    this.app.renderer.on("resize", () => {
      this.app.stage.hitArea = this.app.screen;
    });

    this.initialized = true;
  }

  destroy(): void {
    if (!this.initialized) return;
    this.app.canvas.removeEventListener("wheel", this.handleWheel);
    this.app.canvas.removeEventListener("contextmenu", this.handleContextMenu);
    this.app.destroy(true, { children: true });
    this.initialized = false;
  }

  update(snapshot: BoardSnapshot): void {
    this.snapshot = snapshot;
    this.draw();
  }

  // Called every simulation step (and with null when playback ends/hasn't
  // started) to update the visited/trail/agent overlay without touching the
  // board layers, which only change when the board itself is edited.
  updateSimulation(visual: SimulationVisual | null): void {
    this.simulation = visual;
    this.drawVisited();
    this.drawTrail();
    this.drawAgents();
  }

  // Toggling out of edit mode cancels any in-progress paint stroke and clears hover.
  setEditable(editable: boolean): void {
    this.editable = editable;
    if (!editable) {
      this.paintMode = null;
      this.paintedIds.clear();
      if (this.hovered.kind !== "none") {
        this.hovered = { kind: "none" };
        this.drawHover();
      }
    }
  }

  setTool(tool: EditTool): void {
    this.tool = tool;
    this.paintMode = null;
    this.paintedIds.clear();
  }

  zoomIn = (): void => {
    this.zoomAt(this.app.screen.width / 2, this.app.screen.height / 2, 1.2);
  };

  zoomOut = (): void => {
    this.zoomAt(this.app.screen.width / 2, this.app.screen.height / 2, 1 / 1.2);
  };

  resetView = (): void => {
    this.scale = 1;
    this.offsetX = 0;
    this.offsetY = 0;
    this.applyTransform();
  };

  fitToGrid = (): void => {
    const { cols, rows } = this.snapshot;
    if (cols === 0 || rows === 0) return;

    const gridWidth = cols * CELL_SIZE;
    const gridHeight = rows * CELL_SIZE;
    const padding = 40;
    const availableW = Math.max(this.app.screen.width - padding * 2, 10);
    const availableH = Math.max(this.app.screen.height - padding * 2, 10);
    const fitScale = Math.min(availableW / gridWidth, availableH / gridHeight, MAX_SCALE);

    this.scale = Math.max(fitScale, MIN_SCALE);
    this.offsetX = (this.app.screen.width - gridWidth * this.scale) / 2;
    this.offsetY = (this.app.screen.height - gridHeight * this.scale) / 2;
    this.applyTransform();
  };

  private applyTransform(): void {
    this.world.position.set(this.offsetX, this.offsetY);
    this.world.scale.set(this.scale);
  }

  private zoomAt(screenX: number, screenY: number, factor: number): void {
    const worldX = (screenX - this.offsetX) / this.scale;
    const worldY = (screenY - this.offsetY) / this.scale;
    const newScale = Math.min(Math.max(this.scale * factor, MIN_SCALE), MAX_SCALE);

    this.offsetX = screenX - worldX * newScale;
    this.offsetY = screenY - worldY * newScale;
    this.scale = newScale;
    this.applyTransform();
  }

  private handleContextMenu = (e: Event): void => {
    e.preventDefault();
  };

  private handleWheel = (e: WheelEvent): void => {
    e.preventDefault();
    const rect = this.app.canvas.getBoundingClientRect();
    const screenX = e.clientX - rect.left;
    const screenY = e.clientY - rect.top;
    const factor = e.deltaY < 0 ? 1.1 : 1 / 1.1;
    this.zoomAt(screenX, screenY, factor);
  };

  private handlePointerDown = (e: FederatedPointerEvent): void => {
    // Right button: pan only, never edits the board.
    if (e.button === 2) {
      this.isPanning = true;
      this.panStartScreen = { x: e.global.x, y: e.global.y };
      this.panStartOffset = { x: this.offsetX, y: this.offsetY };
      return;
    }

    if (e.button !== 0 || !this.editable) return;

    // Left button: start a paint stroke. The kind hit first (vertex or edge)
    // locks the stroke; the resulting on/off value is re-applied to every new
    // target the pointer passes over until release.
    if (this.tool === "start") {
      const hit = this.hitTest(e.global.x, e.global.y, "vertex");
      if (hit.kind === "vertex" && this.snapshot.activeVertices.has(hit.id)) {
        this.paintMode = "start";
        this.paintValue = !this.snapshot.startPoints.has(hit.id);
        this.paintedIds = new Set([hit.id]);
        this.callbacks.onStartPointPaint(hit.id, this.paintValue);
      }
      return;
    }

    const hit = this.hitTest(e.global.x, e.global.y);
    if (hit.kind === "vertex") {
      if (this.snapshot.startPoints.has(hit.id)) {
        // A cell carrying a start-point marker loses the marker first; the
        // cell itself is only removed on a subsequent click/stroke.
        this.paintMode = "eraseStart";
        this.paintedIds = new Set([hit.id]);
        this.callbacks.onStartPointPaint(hit.id, false);
      } else {
        this.paintMode = "vertex";
        this.paintValue = !this.snapshot.activeVertices.has(hit.id);
        this.paintedIds = new Set([hit.id]);
        this.callbacks.onVertexPaint(hit.id, this.paintValue);
      }
    } else if (hit.kind === "edge") {
      this.paintMode = "edge";
      this.paintValue = !this.snapshot.walls.has(hit.key);
      this.paintedIds = new Set([hit.key]);
      this.callbacks.onEdgePaint(hit.key, hit.a, hit.b, this.paintValue);
    }
  };

  private handlePointerMove = (e: FederatedPointerEvent): void => {
    const screenX = e.global.x;
    const screenY = e.global.y;

    if (this.isPanning) {
      const dx = screenX - this.panStartScreen.x;
      const dy = screenY - this.panStartScreen.y;
      this.offsetX = this.panStartOffset.x + dx;
      this.offsetY = this.panStartOffset.y + dy;
      this.applyTransform();
      return;
    }

    if (!this.editable) {
      if (this.hovered.kind !== "none") {
        this.hovered = { kind: "none" };
        this.drawHover();
      }
      return;
    }

    if (this.paintMode === "start" || this.paintMode === "eraseStart") {
      const hit = this.hitTest(screenX, screenY, "vertex");
      if (hit.kind === "vertex" && !this.paintedIds.has(hit.id)) {
        if (this.paintMode === "eraseStart") {
          this.paintedIds.add(hit.id);
          this.callbacks.onStartPointPaint(hit.id, false);
        } else if (this.snapshot.activeVertices.has(hit.id)) {
          this.paintedIds.add(hit.id);
          this.callbacks.onStartPointPaint(hit.id, this.paintValue);
        }
      }
      if (!hitEquals(hit, this.hovered)) {
        this.hovered = hit;
        this.drawHover();
      }
      return;
    }

    if (this.paintMode) {
      const hit = this.hitTest(screenX, screenY, this.paintMode);
      if (hit.kind === "vertex" && !this.paintedIds.has(hit.id)) {
        this.paintedIds.add(hit.id);
        this.callbacks.onVertexPaint(hit.id, this.paintValue);
      } else if (hit.kind === "edge" && !this.paintedIds.has(hit.key)) {
        this.paintedIds.add(hit.key);
        this.callbacks.onEdgePaint(hit.key, hit.a, hit.b, this.paintValue);
      }
      if (!hitEquals(hit, this.hovered)) {
        this.hovered = hit;
        this.drawHover();
      }
      return;
    }

    const hit = this.hitTest(screenX, screenY, this.tool === "start" ? "vertex" : undefined);
    if (!hitEquals(hit, this.hovered)) {
      this.hovered = hit;
      this.drawHover();
    }
  };

  private handlePointerUp = (): void => {
    this.isPanning = false;
    this.paintMode = null;
    this.paintedIds.clear();
  };

  // `restrict` locks the search to one kind of target, used while a paint
  // stroke is in progress so the cursor can't switch from painting edges to
  // painting vertices (or vice versa) mid-drag.
  private hitTest(screenX: number, screenY: number, restrict?: "vertex" | "edge"): HitResult {
    const worldX = (screenX - this.offsetX) / this.scale;
    const worldY = (screenY - this.offsetY) / this.scale;
    const { cols, rows, activeVertices } = this.snapshot;

    const col = Math.floor(worldX / CELL_SIZE);
    const row = Math.floor(worldY / CELL_SIZE);
    if (col < 0 || row < 0 || col >= cols || row >= rows) return { kind: "none" };

    const localX = worldX - col * CELL_SIZE;
    const localY = worldY - row * CELL_SIZE;
    const id = `${col},${row}`;

    // Only treat the pointer as being over an edge when both neighboring cells
    // are already part of the board — otherwise fall back to the cell itself.
    // Each border's hitbox is also excluded near its two corners so it doesn't
    // bleed into the perpendicular border (e.g. dragging along a vertical
    // wall no longer occasionally "catches" the horizontal wall it meets).
    if (restrict !== "vertex") {
      const candidates: Array<[number, "left" | "right" | "top" | "bottom"]> = [];
      if (localY >= EDGE_BAND && localY <= CELL_SIZE - EDGE_BAND) {
        candidates.push([localX, "left"], [CELL_SIZE - localX, "right"]);
      }
      if (localX >= EDGE_BAND && localX <= CELL_SIZE - EDGE_BAND) {
        candidates.push([localY, "top"], [CELL_SIZE - localY, "bottom"]);
      }

      if (candidates.length > 0) {
        candidates.sort((a, b) => a[0] - b[0]);
        const [minDist, side] = candidates[0];

        if (minDist < EDGE_BAND) {
          let ncol = col;
          let nrow = row;
          if (side === "left") ncol -= 1;
          else if (side === "right") ncol += 1;
          else if (side === "top") nrow -= 1;
          else nrow += 1;

          if (ncol >= 0 && nrow >= 0 && ncol < cols && nrow < rows) {
            const neighborId = `${ncol},${nrow}`;
            if (activeVertices.has(id) && activeVertices.has(neighborId)) {
              return { kind: "edge", key: edgeKey(col, row, ncol, nrow), a: id, b: neighborId };
            }
          }
        }
      }
    }

    if (restrict === "edge") return { kind: "none" };

    return { kind: "vertex", id };
  }

  private draw(): void {
    this.drawGrid();
    this.drawCells();
    this.drawVisited();
    this.drawWalls();
    this.drawTrail();
    this.drawStartPoints();
    this.drawAgents();
    this.drawHover();
  }

  private drawGrid(): void {
    const { cols, rows } = this.snapshot;
    this.gridLayer.clear();
    if (cols === 0 || rows === 0) return;

    const width = cols * CELL_SIZE;
    const height = rows * CELL_SIZE;

    for (let x = 0; x <= cols; x++) {
      this.gridLayer.moveTo(x * CELL_SIZE, 0).lineTo(x * CELL_SIZE, height);
    }
    for (let y = 0; y <= rows; y++) {
      this.gridLayer.moveTo(0, y * CELL_SIZE).lineTo(width, y * CELL_SIZE);
    }
    this.gridLayer.stroke({ width: 1, color: COLOR_GRID_LINE });
  }

  private drawCells(): void {
    this.cellsLayer.clear();
    const { activeVertices } = this.snapshot;
    if (activeVertices.size === 0) return;

    for (const id of activeVertices) {
      const [x, y] = parseVertexId(id);
      this.cellsLayer.rect(x * CELL_SIZE, y * CELL_SIZE, CELL_SIZE, CELL_SIZE);
    }
    this.cellsLayer.fill({ color: COLOR_CELL_ACTIVE, alpha: 0.55 });
  }

  private drawWalls(): void {
    this.wallsLayer.clear();
    const { walls } = this.snapshot;
    if (walls.size === 0) return;

    for (const key of walls) {
      const [a, b] = edgeEndpoints(key);
      const rect = edgeRect(a, b, WALL_THICKNESS);
      this.wallsLayer.rect(rect.x, rect.y, rect.w, rect.h);
    }
    this.wallsLayer.fill({ color: COLOR_WALL });
  }

  private drawStartPoints(): void {
    this.startLayer.clear();
    const { startPoints } = this.snapshot;
    if (startPoints.size === 0) return;

    for (const id of startPoints) {
      const [x, y] = parseVertexId(id);
      const cx = x * CELL_SIZE + CELL_SIZE / 2;
      const cy = y * CELL_SIZE + CELL_SIZE / 2;
      this.startLayer.circle(cx, cy, START_RADIUS);
    }
    this.startLayer.fill({ color: COLOR_START });
  }

  // Cumulative "heatmap" of every cell any agent has stepped on so far: color
  // is the average of every visiting agent's color (a cheap, good-enough color
  // fusion for overlapping routes), alpha ramps up with repeat visits.
  private drawVisited(): void {
    this.visitLayer.clear();
    if (!this.simulation) return;

    for (const [id, { color, count }] of this.simulation.visited) {
      const [x, y] = parseVertexId(id);
      const alpha = Math.min(VISIT_MAX_ALPHA, VISIT_BASE_ALPHA + VISIT_ALPHA_STEP * (count - 1));
      this.visitLayer.rect(x * CELL_SIZE, y * CELL_SIZE, CELL_SIZE, CELL_SIZE).fill({ color, alpha });
    }
  }

  // Each segment is stroked independently so overlapping trails from
  // different agents blend via alpha compositing instead of one hiding
  // another.
  private drawTrail(): void {
    this.trailLayer.clear();
    if (!this.simulation) return;

    for (const segment of this.simulation.trail) {
      const [ax, ay] = parseVertexId(segment.from);
      const [bx, by] = parseVertexId(segment.to);
      const startX = ax * CELL_SIZE + CELL_SIZE / 2;
      const startY = ay * CELL_SIZE + CELL_SIZE / 2;
      const endX = bx * CELL_SIZE + CELL_SIZE / 2;
      const endY = by * CELL_SIZE + CELL_SIZE / 2;
      this.trailLayer
        .moveTo(startX, startY)
        .lineTo(endX, endY)
        .stroke({ width: TRAIL_WIDTH, color: segment.color, alpha: TRAIL_ALPHA, cap: "round" });
    }
  }

  // Agents sharing a cell are arranged in a small ring instead of stacking
  // exactly on top of each other, so every one of them stays visible.
  private drawAgents(): void {
    this.agentLayer.clear();
    if (!this.simulation) return;

    const byVertex = new Map<string, Array<{ id: number; color: number }>>();
    for (const agent of this.simulation.agents) {
      const list = byVertex.get(agent.vertex);
      if (list) list.push(agent);
      else byVertex.set(agent.vertex, [agent]);
    }

    for (const [vertexId, agents] of byVertex) {
      const [x, y] = parseVertexId(vertexId);
      const cx = x * CELL_SIZE + CELL_SIZE / 2;
      const cy = y * CELL_SIZE + CELL_SIZE / 2;

      const clustered = agents.length > 1;
      const radius = clustered ? AGENT_RADIUS * 0.65 : AGENT_RADIUS;
      const offset = clustered ? AGENT_RADIUS * 0.7 : 0;

      agents.forEach((agent, i) => {
        const angle = (2 * Math.PI * i) / agents.length;
        const px = cx + offset * Math.cos(angle);
        const py = cy + offset * Math.sin(angle);
        this.agentLayer
          .circle(px, py, radius)
          .fill({ color: agent.color })
          .stroke({ width: 1.5, color: AGENT_STROKE_COLOR, alpha: 0.6 });
      });
    }
  }

  private drawHover(): void {
    this.hoverLayer.clear();
    const hit = this.hovered;

    if (hit.kind === "vertex") {
      const [x, y] = parseVertexId(hit.id);
      this.hoverLayer.rect(x * CELL_SIZE, y * CELL_SIZE, CELL_SIZE, CELL_SIZE).fill({ color: COLOR_HOVER, alpha: 0.12 });
    } else if (hit.kind === "edge") {
      const rect = edgeRect(hit.a, hit.b, WALL_THICKNESS);
      this.hoverLayer.rect(rect.x, rect.y, rect.w, rect.h).fill({ color: COLOR_HOVER, alpha: 0.25 });
    }
  }
}
