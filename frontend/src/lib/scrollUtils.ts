export const SCROLL_INTO_VIEW_DELAY_MS = 100

export function hasScrollIntoView(element: Element): element is Element & { scrollIntoView: (options?: ScrollIntoViewOptions) => void } {
  return typeof (element as any).scrollIntoView === 'function'
}

export function scrollToElement(
  selector: string,
  options: {
    delay?: number
    behavior?: ScrollBehavior
    block?: ScrollLogicalPosition
  } = {}
): void {
  const {
    delay = SCROLL_INTO_VIEW_DELAY_MS,
    behavior = 'smooth',
    block = 'center'
  } = options

  setTimeout(() => {
    const element = document.querySelector(selector)
    if (element && hasScrollIntoView(element)) {
      element.scrollIntoView({ behavior, block })
    }
  }, delay)
}

export function scrollToPosition(
  position: number,
  options: {
    delay?: number
    behavior?: ScrollBehavior
  } = {}
): void {
  const {
    delay = SCROLL_INTO_VIEW_DELAY_MS,
    behavior = 'smooth'
  } = options

  const scrollFn = () => {
    window.scrollTo({ top: position, behavior })
  }

  if (delay > 0) {
    requestAnimationFrame(() => {
      setTimeout(scrollFn, delay)
    })
  } else {
    requestAnimationFrame(scrollFn)
  }
}
