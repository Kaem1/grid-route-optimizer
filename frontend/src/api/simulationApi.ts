import { apiGet, apiPost } from "./client";

// Mirrors backend/internal/api/dto.go — kept in sync manually since the
// project has no shared schema codegen yet.
export interface GridDTO {
  cols: number;
  rows: number;
  activeVertices: string[];
  walls: string[];
}

export interface RouteDTO {
  startPoint: string;
  vertices: string[];
  length: number;
}

export interface MetricsDTO {
  totalLength: number;
  numberOfRoutes: number;
  coveredVertices: number;
  coverageRatio: number;
}

export interface SimulateResponse {
  algorithm: string;
  routes: RouteDTO[];
  metrics: MetricsDTO;
}

export interface AlgorithmsResponse {
  algorithms: string[];
}

export function getAlgorithms(): Promise<AlgorithmsResponse> {
  return apiGet<AlgorithmsResponse>("/api/algorithms");
}

export function runSimulation(grid: GridDTO, startPoints: string[], algorithm: string): Promise<SimulateResponse> {
  return apiPost<SimulateResponse>("/api/simulate", { grid, startPoints, algorithm });
}
