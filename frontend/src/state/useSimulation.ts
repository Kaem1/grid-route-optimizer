import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { MetricsDTO, RouteDTO, SimulateResponse } from "../api/simulationApi";
import { agentColor, packRGB, randomHueOffset, unpackRGB } from "../utils/color";

export interface SimulationAgent {
  id: number;
  vertex: string;
  color: number;
}

export interface SimulationVisited {
  color: number;
  count: number;
}

export interface SimulationSnapshot {
  visited: ReadonlyMap<string, SimulationVisited>;
  trail: ReadonlyArray<{ from: string; to: string; color: number }>;
  agents: ReadonlyArray<SimulationAgent>;
}

export interface SimulationState {
  active: boolean;
  algorithm: string;
  metrics: MetricsDTO | null;
  step: number;
  maxStep: number;
  isPlaying: boolean;
  speed: number;
  snapshot: SimulationSnapshot;
  start: (response: SimulateResponse) => void;
  stop: () => void;
  stepForward: () => void;
  stepBackward: () => void;
  togglePlay: () => void;
  setSpeed: (speed: number) => void;
}

// One agent's movement at a given simulation step: `from` is null only for
// the very first step, where the agent is simply placed at its start point.
interface StepEvent {
  agentIndex: number;
  from: string | null;
  to: string;
}

interface TrailSegment {
  step: number;
  from: string;
  to: string;
  color: number;
}

interface Runtime {
  routes: string[][];
  colors: number[];
  stepEvents: StepEvent[][];
  maxStep: number;
  visitCounts: Map<string, number>;
  visitColorSum: Map<string, [number, number, number]>;
  trail: TrailSegment[];
}

const EMPTY_SNAPSHOT: SimulationSnapshot = { visited: new Map(), trail: [], agents: [] };
const MIN_SPEED = 1;
const MAX_SPEED = 10;
const DEFAULT_SPEED = 3;

// Precomputes, for every step, which agents actually move and to where.
// Routes may finish before the longest one does — those agents simply stay
// put (no event) once their walk is exhausted.
function buildStepEvents(routes: RouteDTO[]): StepEvent[][] {
  const maxStep = routes.reduce((max, route) => Math.max(max, route.vertices.length - 1), 0);
  const events: StepEvent[][] = Array.from({ length: maxStep + 1 }, () => []);

  routes.forEach((route, agentIndex) => {
    for (let step = 0; step <= maxStep; step++) {
      const toIndex = Math.min(step, route.vertices.length - 1);
      const fromIndex = step === 0 ? -1 : Math.min(step - 1, route.vertices.length - 1);
      const to = route.vertices[toIndex];
      const from = fromIndex === -1 ? null : route.vertices[fromIndex];
      if (from === to) continue; // route already finished (or a same-cell no-op), agent stays put
      events[step].push({ agentIndex, from, to });
    }
  });

  return events;
}

// Mutates the runtime's accumulated visit/trail state by one step, in
// either direction (sign = +1 applies the step, -1 undoes it).
function applyStepEvents(runtime: Runtime, step: number, sign: 1 | -1): void {
  for (const event of runtime.stepEvents[step] ?? []) {
    const color = runtime.colors[event.agentIndex];
    const [r, g, b] = unpackRGB(color);

    const prevCount = runtime.visitCounts.get(event.to) ?? 0;
    const nextCount = prevCount + sign;
    if (nextCount <= 0) {
      runtime.visitCounts.delete(event.to);
      runtime.visitColorSum.delete(event.to);
    } else {
      runtime.visitCounts.set(event.to, nextCount);
      const [sr, sg, sb] = runtime.visitColorSum.get(event.to) ?? [0, 0, 0];
      runtime.visitColorSum.set(event.to, [sr + r * sign, sg + g * sign, sb + b * sign]);
    }

    if (event.from !== null) {
      if (sign > 0) {
        runtime.trail.push({ step, from: event.from, to: event.to, color });
      } else {
        runtime.trail = runtime.trail.filter((segment) => segment.step !== step);
      }
    }
  }
}

// Drives the step-by-step playback of a computed solution: precomputes the
// per-step movement events once, then advances/rewinds by mutating a small
// accumulator so every step transition costs O(agents) regardless of how
// long the simulation is (see project instructions section 48 — must stay
// usable on large grids/long routes).
export function useSimulation(): SimulationState {
  const [active, setActive] = useState(false);
  const [algorithm, setAlgorithm] = useState("");
  const [metrics, setMetrics] = useState<MetricsDTO | null>(null);
  const [step, setStep] = useState(0);
  const [maxStep, setMaxStep] = useState(0);
  const [isPlaying, setIsPlaying] = useState(false);
  const [speed, setSpeedState] = useState(DEFAULT_SPEED);
  // Bumped on every start() so the snapshot memo below always recomputes,
  // even though `step` resets to a value (0) it may already hold.
  const [runId, setRunId] = useState(0);

  const runtimeRef = useRef<Runtime | null>(null);
  // Mirrors `step` synchronously: setState's functional updater runs
  // asynchronously (and is double-invoked under Strict Mode), so it can't be
  // used to compute advance()'s return value right after calling it.
  const stepRef = useRef(0);

  const start = useCallback((response: SimulateResponse) => {
    const hueOffset = randomHueOffset();
    const colors = response.routes.map((_, i) => agentColor(i, hueOffset));
    const stepEvents = buildStepEvents(response.routes);
    const runtime: Runtime = {
      routes: response.routes.map((route) => route.vertices),
      colors,
      stepEvents,
      maxStep: stepEvents.length - 1,
      visitCounts: new Map(),
      visitColorSum: new Map(),
      trail: [],
    };
    applyStepEvents(runtime, 0, 1); // seed the initial agent placements

    runtimeRef.current = runtime;
    stepRef.current = 0;
    setRunId((id) => id + 1);
    setAlgorithm(response.algorithm);
    setMetrics(response.metrics);
    setMaxStep(runtime.maxStep);
    setStep(0);
    setIsPlaying(false);
    setActive(true);
  }, []);

  const stop = useCallback(() => {
    runtimeRef.current = null;
    stepRef.current = 0;
    setActive(false);
    setIsPlaying(false);
    setAlgorithm("");
    setMetrics(null);
    setStep(0);
    setMaxStep(0);
  }, []);

  // Shared by the "next step" button and the autoplay timer. Returns
  // whether the simulation actually advanced (false once at the end).
  const advance = useCallback((): boolean => {
    const runtime = runtimeRef.current;
    if (!runtime || stepRef.current >= runtime.maxStep) return false;

    const next = stepRef.current + 1;
    applyStepEvents(runtime, next, 1);
    stepRef.current = next;
    setStep(next);
    return true;
  }, []);

  const stepForward = useCallback(() => {
    advance();
  }, [advance]);

  const stepBackward = useCallback(() => {
    const runtime = runtimeRef.current;
    if (!runtime || stepRef.current <= 0) return;

    applyStepEvents(runtime, stepRef.current, -1);
    stepRef.current -= 1;
    setStep(stepRef.current);
  }, []);

  const togglePlay = useCallback(() => {
    setIsPlaying((playing) => !playing);
  }, []);

  const setSpeed = useCallback((value: number) => {
    setSpeedState(Math.min(MAX_SPEED, Math.max(MIN_SPEED, value)));
  }, []);

  // Autoplay: ticks forward on an interval derived from `speed`, pausing
  // itself once the last step is reached.
  useEffect(() => {
    if (!isPlaying) return;

    const intervalMs = Math.max(1000 / speed, 20);
    const id = window.setInterval(() => {
      if (!advance()) setIsPlaying(false);
    }, intervalMs);

    return () => window.clearInterval(id);
  }, [isPlaying, speed, advance]);

  const snapshot = useMemo<SimulationSnapshot>(() => {
    const runtime = runtimeRef.current;
    if (!runtime) return EMPTY_SNAPSHOT;

    const visited = new Map<string, SimulationVisited>();
    for (const [vertex, count] of runtime.visitCounts) {
      const [r, g, b] = runtime.visitColorSum.get(vertex) ?? [0, 0, 0];
      visited.set(vertex, { color: packRGB(r / count, g / count, b / count), count });
    }

    const agents = runtime.routes.map((route, id) => ({
      id,
      vertex: route[Math.min(step, route.length - 1)],
      color: runtime.colors[id],
    }));

    return { visited, trail: runtime.trail, agents };
    // `step` covers step-by-step playback; `runId` forces a recompute when a
    // new simulation starts even if `step` happens to already read 0.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [step, runId]);

  return {
    active,
    algorithm,
    metrics,
    step,
    maxStep,
    isPlaying,
    speed,
    snapshot,
    start,
    stop,
    stepForward,
    stepBackward,
    togglePlay,
    setSpeed,
  };
}
