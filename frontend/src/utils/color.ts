// Small HSL/RGB helpers used to generate pleasant agent colors and to blend
// overlapping visit colors on the board. Pure math, no rendering deps.

const GOLDEN_ANGLE = 137.508;
const AGENT_SATURATION = 55;
const AGENT_LIGHTNESS = 62;

// Spreads agent hues using the golden angle from a random starting offset:
// looks "randomized" per run while keeping consecutive agents visually
// distinct (unlike fully uniform random hues, which often clash or repeat).
export function randomHueOffset(): number {
  return Math.random() * 360;
}

export function agentColor(index: number, hueOffset: number): number {
  const hue = (hueOffset + index * GOLDEN_ANGLE) % 360;
  return hslToPacked(hue, AGENT_SATURATION, AGENT_LIGHTNESS);
}

export function packRGB(r: number, g: number, b: number): number {
  return (clampByte(r) << 16) | (clampByte(g) << 8) | clampByte(b);
}

export function unpackRGB(color: number): [number, number, number] {
  return [(color >> 16) & 0xff, (color >> 8) & 0xff, color & 0xff];
}

function clampByte(value: number): number {
  return Math.max(0, Math.min(255, Math.round(value)));
}

function hslToPacked(h: number, s: number, l: number): number {
  const sat = s / 100;
  const light = l / 100;
  const c = (1 - Math.abs(2 * light - 1)) * sat;
  const hh = h / 60;
  const x = c * (1 - Math.abs((hh % 2) - 1));

  let r = 0;
  let g = 0;
  let b = 0;
  if (hh < 1) [r, g, b] = [c, x, 0];
  else if (hh < 2) [r, g, b] = [x, c, 0];
  else if (hh < 3) [r, g, b] = [0, c, x];
  else if (hh < 4) [r, g, b] = [0, x, c];
  else if (hh < 5) [r, g, b] = [x, 0, c];
  else [r, g, b] = [c, 0, x];

  const m = light - c / 2;
  return packRGB((r + m) * 255, (g + m) * 255, (b + m) * 255);
}
