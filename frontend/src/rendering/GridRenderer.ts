import { Application, Container, type FederatedPointerEvent, Graphics } from "pixi.js";
import { edgeEndpoints, edgeKey } from "../graph/grid";

// Visual size of a single cell in world units. Purely a rendering concern.
export const CELL_SIZE = 48;

const EDGE_BAND = 10;
const MIN_SCALE = 0.2;
const MAX_SCALE = 4;
const DRAG_THRESHOLD = 4;
const WALL_THICKNESS = CELL_SIZE * 0.34;

const COLOR_GRID_LINE = 0x333333;
const COLOR_CELL_ACTIVE = 0x2f6fed;
const COLOR_WALL = 0xff5544;
const COLOR_HOVER = 0xffffff;

export interface BoardSnapshot {
  cols: number;
  rows: number;
  activeVertices: ReadonlySet<string>;
  walls: ReadonlySet<string>;
}

export interface GridRendererCallbacks {
  onVertexClick: (id: string) => void;
  onEdgeClick: (key: string, a: string, b: string) => void;
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
  private readonly wallsLayer = new Graphics();
  private readonly hoverLayer = new Graphics();

  private snapshot: BoardSnapshot = { cols: 0, rows: 0, activeVertices: new Set(), walls: new Set() };

  private scale = 1;
  private offsetX = 0;
  private offsetY = 0;

  private isPointerDown = false;
  private hasDragged = false;
  private dragStartScreen = { x: 0, y: 0 };
  private dragStartOffset = { x: 0, y: 0 };

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

    this.world.addChild(this.gridLayer, this.cellsLayer, this.wallsLayer, this.hoverLayer);
    this.app.stage.addChild(this.world);

    this.app.stage.eventMode = "static";
    this.app.stage.hitArea = this.app.screen;

    this.app.stage.on("pointerdown", this.handlePointerDown);
    this.app.stage.on("pointermove", this.handlePointerMove);
    this.app.stage.on("pointerup", this.handlePointerUp);
    this.app.stage.on("pointerupoutside", this.handlePointerUp);
    this.app.canvas.addEventListener("wheel", this.handleWheel, { passive: false });

    this.app.renderer.on("resize", () => {
      this.app.stage.hitArea = this.app.screen;
    });

    this.initialized = true;
  }

  destroy(): void {
    if (!this.initialized) return;
    this.app.canvas.removeEventListener("wheel", this.handleWheel);
    this.app.destroy(true, { children: true });
    this.initialized = false;
  }

  update(snapshot: BoardSnapshot): void {
    this.snapshot = snapshot;
    this.draw();
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

  private handleWheel = (e: WheelEvent): void => {
    e.preventDefault();
    const rect = this.app.canvas.getBoundingClientRect();
    const screenX = e.clientX - rect.left;
    const screenY = e.clientY - rect.top;
    const factor = e.deltaY < 0 ? 1.1 : 1 / 1.1;
    this.zoomAt(screenX, screenY, factor);
  };

  private handlePointerDown = (e: FederatedPointerEvent): void => {
    this.isPointerDown = true;
    this.hasDragged = false;
    this.dragStartScreen = { x: e.global.x, y: e.global.y };
    this.dragStartOffset = { x: this.offsetX, y: this.offsetY };
  };

  private handlePointerMove = (e: FederatedPointerEvent): void => {
    const screenX = e.global.x;
    const screenY = e.global.y;

    if (this.isPointerDown) {
      const dx = screenX - this.dragStartScreen.x;
      const dy = screenY - this.dragStartScreen.y;
      if (!this.hasDragged && Math.hypot(dx, dy) > DRAG_THRESHOLD) {
        this.hasDragged = true;
      }
      if (this.hasDragged) {
        this.offsetX = this.dragStartOffset.x + dx;
        this.offsetY = this.dragStartOffset.y + dy;
        this.applyTransform();
      }
    }

    if (this.hasDragged) {
      if (this.hovered.kind !== "none") {
        this.hovered = { kind: "none" };
        this.drawHover();
      }
      return;
    }

    const hit = this.hitTest(screenX, screenY);
    if (!hitEquals(hit, this.hovered)) {
      this.hovered = hit;
      this.drawHover();
    }
  };

  private handlePointerUp = (e: FederatedPointerEvent): void => {
    if (this.isPointerDown && !this.hasDragged) {
      const hit = this.hitTest(e.global.x, e.global.y);
      if (hit.kind === "vertex") this.callbacks.onVertexClick(hit.id);
      else if (hit.kind === "edge") this.callbacks.onEdgeClick(hit.key, hit.a, hit.b);
    }
    this.isPointerDown = false;
    this.hasDragged = false;
  };

  private hitTest(screenX: number, screenY: number): HitResult {
    const worldX = (screenX - this.offsetX) / this.scale;
    const worldY = (screenY - this.offsetY) / this.scale;
    const { cols, rows, activeVertices } = this.snapshot;

    const col = Math.floor(worldX / CELL_SIZE);
    const row = Math.floor(worldY / CELL_SIZE);
    if (col < 0 || row < 0 || col >= cols || row >= rows) return { kind: "none" };

    const localX = worldX - col * CELL_SIZE;
    const localY = worldY - row * CELL_SIZE;
    const candidates: Array<[number, "left" | "right" | "top" | "bottom"]> = [
      [localX, "left"],
      [CELL_SIZE - localX, "right"],
      [localY, "top"],
      [CELL_SIZE - localY, "bottom"],
    ];
    candidates.sort((a, b) => a[0] - b[0]);
    const [minDist, side] = candidates[0];

    const id = `${col},${row}`;

    // Only treat the pointer as being over an edge when both neighboring cells
    // are already part of the board — otherwise fall back to the cell itself.
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

    return { kind: "vertex", id };
  }

  private draw(): void {
    this.drawGrid();
    this.drawCells();
    this.drawWalls();
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
