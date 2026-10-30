export type VirtualMessage = { id: string; height: number }

export function preserveScrollAnchor(previousTop: number, previousHeight: number, nextHeight: number) {
  return previousTop + nextHeight - previousHeight
}

export function visibleRange(start: number, end: number, overscan: number, count: number) {
  return { start: Math.max(0, start - overscan), end: Math.min(count, end + overscan) }
}
