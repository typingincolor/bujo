# Resizable Journal Sidebar Design

**Date:** 2026-01-25
**Status:** Approved
**Context:** Enable users to drag the journal sidebar edge to adjust width for better content visibility

## Overview

Add resize functionality to the journal sidebar (currently fixed at 32rem/512px) allowing users to drag the left edge to adjust width between 24rem (384px) and 60rem (960px). Width resets to default on app reload.

## Design Decisions

### Interaction Model
- **Live resize**: Width updates immediately as user drags (smooth feedback)
- **Drag target**: Left edge of sidebar with larger hit area for easy targeting
- **Visual feedback**: Cursor changes to col-resize (↔) on hover, subtle hover highlight on edge
- **Persistence**: Width resets to default 32rem on reload (no localStorage)

### Size Constraints
- **Minimum**: 24rem (384px) - prevents sidebar from becoming unusable
- **Maximum**: 60rem (960px) - prevents sidebar from dominating screen
- **Default**: 32rem (512px) - current width, balanced for most use cases

## Component Structure

### JournalSidebar.tsx Changes

**State management:**
```typescript
const [sidebarWidth, setSidebarWidth] = useState(512) // 32rem default
const [isResizing, setIsResizing] = useState(false)
```

**Root element with dynamic width:**
```typescript
<div
  className={cn(
    "flex flex-col h-full relative",
    isResizing && "select-none"
  )}
  style={{ width: `${sidebarWidth}px` }}
>
```

**Resize handle (left edge):**
```typescript
{!isCollapsed && (
  <div
    className="absolute left-0 top-0 h-full w-2 cursor-col-resize hover:bg-primary/10 transition-colors"
    onMouseDown={handleResizeStart}
  />
)}
```

### Resize Interaction Logic

**Three-phase mouse interaction:**

1. **Start (onMouseDown):**
   - Set `isResizing` to true
   - Attach global mousemove and mouseup listeners
   - Prevent default to avoid text selection

2. **Move (global mousemove):**
   - Calculate new width: `windowWidth - mouseX`
   - Clamp between min (384px) and max (960px)
   - Update `sidebarWidth` state immediately

3. **End (global mouseup):**
   - Set `isResizing` to false
   - Remove global listeners

**Cleanup:**
- Remove event listeners on component unmount
- Reset body cursor and user-select styles

### App.tsx Integration

**Track sidebar width in parent:**
```typescript
const [journalSidebarWidth, setJournalSidebarWidth] = useState(512)
```

**Add callback prop to JournalSidebar:**
```typescript
interface JournalSidebarProps {
  // ... existing props
  onWidthChange?: (width: number) => void;
}
```

**Update main content positioning:**
Replace static `right-[32rem]` class with dynamic margin:
```typescript
<div
  style={{
    marginRight: isSidebarCollapsed ? '0' : `${journalSidebarWidth}px`
  }}
  className="flex-1 transition-[margin]"
>
```

## Polish & Edge Cases

### Visual Feedback During Resize
- Apply `select-none` class to sidebar when `isResizing` is true
- Set global `cursor: col-resize` and `user-select: none` on document.body during drag
- Reset body styles when resize ends

### Handle Visibility
- Only show resize handle when sidebar is not collapsed
- Handle should not interfere with collapse toggle button (positioned top-right)

### Performance
- Live width updates use direct state changes (React handles efficiently)
- No debouncing needed - resize is local component state

## Testing Considerations

- Width stays within min (384px) / max (960px) bounds
- Resize handle doesn't interfere with collapse button
- Scrolling works in both sections during and after resize
- Smooth live updates without performance issues
- Cursor changes correctly on hover and during drag
- Event listeners cleaned up properly on unmount
- Width resets to 512px on app reload

## Implementation Order

1. Add resize state and handlers to JournalSidebar.tsx
2. Add resize handle element with hover styles
3. Implement mouse event handlers (start, move, end)
4. Add cleanup useEffect for event listeners
5. Add global cursor/select styles during resize
6. Add onWidthChange prop and callback
7. Update App.tsx to track and use dynamic sidebar width
8. Test all edge cases
9. Update tests to verify resize behavior
