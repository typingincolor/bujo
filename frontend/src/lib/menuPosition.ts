interface Point {
  x: number
  y: number
}

interface Size {
  width: number
  height: number
}

export function calculateMenuPosition(
  clickPos: Point,
  menuSize: Size,
  viewport: Size
): Point {
  let x = clickPos.x
  let y = clickPos.y

  if (x + menuSize.width > viewport.width) {
    x = viewport.width - menuSize.width
  }

  if (y + menuSize.height > viewport.height) {
    y = clickPos.y - menuSize.height
  }

  x = Math.max(0, x)
  y = Math.max(0, y)

  return { x, y }
}
