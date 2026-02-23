# Replace JSON.parse(JSON.stringify(...)) with structuredClone()

Replace 4 deep-copy idioms in `DashboardView.jsx` with the native `structuredClone()` API for improved performance and clarity.

---

## Task 1: Replace Deep-Copy Idiom with structuredClone()

**Files:** `frontend/src/components/DashboardView.jsx`
**Context:** The component uses `JSON.parse(JSON.stringify(obj))` in four places to deep-clone widget arrays before pushing them onto undo/redo stacks or into layout state. `structuredClone()` is a native browser API (baseline since March 2022, supported in all modern browsers and Node 17+) that is more performant, handles more types correctly, and communicates intent more clearly. The widget objects are simple serializable data, so both approaches produce identical results — but `structuredClone()` is the idiomatic modern choice.

**Step-by-Step:**

1. **Replace deep copy in `pushUndo` (line 49):**
   Replace `JSON.parse(JSON.stringify(currentWidgets))` with `structuredClone(currentWidgets)`.

2. **Replace deep copy in `undo` (line 60):**
   Replace `JSON.parse(JSON.stringify(widgets()))` with `structuredClone(widgets())`.

3. **Replace deep copy in `redo` (line 72):**
   Replace `JSON.parse(JSON.stringify(widgets()))` with `structuredClone(widgets())`.

4. **Replace deep copy in `handleSaveLayout` (line 237):**
   Replace `JSON.parse(JSON.stringify(widgets()))` with `structuredClone(widgets())`.

5. **Replace deep copy in `handleLoadLayout` (line 252):**
   Replace `JSON.parse(JSON.stringify(layout.widgets))` with `structuredClone(layout.widgets)`.

**Acceptance Criteria:**

* [ ] Zero occurrences of `JSON.parse(JSON.stringify(` remain in `DashboardView.jsx`.
* [ ] All 5 call sites use `structuredClone()` instead.
* [ ] Existing tests in `DashboardView.test.jsx` continue to pass (`npm test`).
* [ ] No new imports are needed (`structuredClone` is a global).
