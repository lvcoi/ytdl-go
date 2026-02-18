# Specification: Dashboard Evolution v2 P0

## Overview
Transform the dashboard into a modular grid system that allows users to customize their layout. This includes adding a high-visibility Direct Download module and implementing "Edit Mode" for layout persistence and drag-and-drop reordering.

## P0 Requirements
1. **Modular Dashboard Grid:** The dashboard content should be organized into a flexible grid of components (modules).
2. **Direct Download Module:** A prominent, single-line input field on the dashboard for immediate download triggering.
3. **Edit Mode:** A toggleable state that allows users to add, remove, and potentially reorder modules.
4. **Layout Persistence:** The user's dashboard configuration must be saved to `localStorage`.
5. **UI Consistency:** Header alignment must be perfect, and the overall aesthetic must match the "Media-First" vibrant theme.

## Technical Constraints
- Use SolidJS for reactivity.
- Use Tailwind CSS for styling.
- Ensure all new components are fully tested.
- Use `batch()` updates for high-frequency SSE data to prevent UI locks.
- Map backend `done` status to UI `complete` status.
