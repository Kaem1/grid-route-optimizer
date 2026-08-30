// Pure grid/graph helpers. No rendering or React dependencies here (see project instructions, section 19/65).

export function vertexId(x: number, y: number): string {
  return `${x},${y}`;
}

// Canonical, direction-independent key for an edge between two adjacent vertices.
export function edgeKey(x1: number, y1: number, x2: number, y2: number): string {
  const a = vertexId(x1, y1);
  const b = vertexId(x2, y2);
  return a < b ? `${a}|${b}` : `${b}|${a}`;
}

export function edgeEndpoints(key: string): [string, string] {
  const [a, b] = key.split("|");
  return [a, b];
}

export function edgeTouchesVertex(key: string, id: string): boolean {
  const [a, b] = edgeEndpoints(key);
  return a === id || b === id;
}
