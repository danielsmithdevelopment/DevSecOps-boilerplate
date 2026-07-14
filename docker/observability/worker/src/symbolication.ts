import type { StackFrame } from './types';

interface SourceMapV3 {
  version: number;
  sources: string[];
  names: string[];
  mappings: string;
  sourcesContent?: string[];
}

/**
 * Minimal VLQ decoder for source-map mappings.
 * Prefer Alloy-pipeline symbolication for high-volume deployments; Worker-side
 * symbolication is for lower-volume / simpler setups.
 */
function decodeVLQ(segment: string): number[] {
  const result: number[] = [];
  let shift = 0;
  let value = 0;
  for (let i = 0; i < segment.length; i++) {
    let digit = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/'.indexOf(
      segment[i]!,
    );
    if (digit === -1) continue;
    const hasContinuation = (digit & 32) !== 0;
    digit &= 31;
    value += digit << shift;
    if (hasContinuation) {
      shift += 5;
      continue;
    }
    const shouldNegate = (value & 1) === 1;
    value >>= 1;
    result.push(shouldNegate ? -value : value);
    value = 0;
    shift = 0;
  }
  return result;
}

/**
 * Map a single generated (minified) stack frame back to original source
 * using a Source Map v3 object. Returns the original frame when mapping fails.
 */
export function symbolicateFrame(frame: StackFrame, map: SourceMapV3): StackFrame {
  const line = frame.lineno;
  const column = frame.colno;
  if (!line || column === undefined || !map.mappings) {
    return frame;
  }

  const lines = map.mappings.split(';');
  const targetLine = line - 1;
  if (targetLine < 0 || targetLine >= lines.length) {
    return frame;
  }

  let genCol = 0;
  let sourceIndex = 0;
  let origLine = 0;
  let origCol = 0;
  let nameIndex = 0;
  let best: { sourceIndex: number; origLine: number; origCol: number; nameIndex: number } | null =
    null;

  const segments = lines[targetLine]!.split(',');
  for (const segment of segments) {
    if (!segment) continue;
    const vals = decodeVLQ(segment);
    if (vals.length === 0) continue;
    genCol += vals[0]!;
    if (vals.length >= 4) {
      sourceIndex += vals[1]!;
      origLine += vals[2]!;
      origCol += vals[3]!;
      if (vals.length >= 5) nameIndex += vals[4]!;
    }
    if (genCol <= column) {
      best = { sourceIndex, origLine, origCol, nameIndex };
    } else {
      break;
    }
  }

  if (!best) return frame;

  return {
    ...frame,
    filename: map.sources[best.sourceIndex] ?? frame.filename,
    lineno: best.origLine + 1,
    colno: best.origCol,
    function: map.names[best.nameIndex] ?? frame.function,
  };
}

/**
 * Fetch a release source map from R2: `{release}/{filename}.map`
 */
export async function fetchSourceMap(
  bucket: R2Bucket,
  release: string,
  filename: string,
): Promise<SourceMapV3 | null> {
  const key = `${release}/${filename.split('/').pop()}.map`;
  const obj = await bucket.get(key);
  if (!obj) return null;
  try {
    return (await obj.json()) as SourceMapV3;
  } catch {
    return null;
  }
}
