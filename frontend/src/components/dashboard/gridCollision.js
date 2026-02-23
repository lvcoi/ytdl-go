// Pure utility functions for dashboard grid collision and layout logic

/**
 * Checks if two widget rectangles overlap.
 * Widgets are considered overlapping if their bounding boxes intersect.
 * @param {object} a Widget A {x, y, width, height}
 * @param {object} b Widget B {x, y, width, height}
 * @returns {boolean} True if they overlap
 */
export function widgetsOverlap(a, b) {
    if (a.id === b.id) return false;
    return (
        a.x < b.x + b.width &&
        a.x + a.width > b.x &&
        a.y < b.y + b.height &&
        a.y + a.height > b.y
    );
}

/**
 * Sorts widgets by their vertical position (top-to-bottom), then horizontal (left-to-right).
 * Useful for processing layout compaction.
 * @param {Array} widgets Array of widget objects
 * @returns {Array} Sorted copy of widgets array
 */
export function sortWidgets(widgets) {
    return [...widgets].sort((a, b) => {
        if (a.y === b.y) return a.x - b.x;
        return a.y - b.y;
    });
}

/**
 * Resolves collisions by pushing overlapping widgets downward.
 * This implements the "Grafana-style" push behavior where moving a widget
 * displaces others down, cascading the effect.
 * 
 * @param {Array} allWidgets Current state of all widgets
 * @param {object} movedWidget The widget that was moved/resized (must be part of allWidgets or have same ID)
 * @returns {Array} New array of widgets with collisions resolved
 */
export function resolveCollisions(allWidgets, movedWidget) {
    // Create a working copy
    let widgets = allWidgets.map(w => w.id === movedWidget.id ? { ...movedWidget } : { ...w });
    
    // Sort by Y to process top-down, but ensure movedWidget is processed first/separately
    // in the recursion logic.
    // However, for the push logic, we just need to find what movedWidget overlaps with.
    
    const pushWidget = (pusher, remainingWidgets) => {
        let modified = false;
        
        remainingWidgets.forEach(target => {
            if (target.id === pusher.id) return;
            if (!target.enabled) return; // Ignore disabled widgets
            
            if (widgetsOverlap(pusher, target)) {
                // Collision detected! Push target down.
                // Move target to just below the pusher
                target.y = pusher.y + pusher.height;
                modified = true;
                
                // Recursively push anything that the target now overlaps
                pushWidget(target, remainingWidgets);
            }
        });
        
        return modified;
    };

    // Start the cascade from the moved widget
    const moved = widgets.find(w => w.id === movedWidget.id);
    if (moved && moved.enabled) {
        pushWidget(moved, widgets);
    }

    return widgets;
}

/**
 * Compacts the layout by moving widgets up to fill empty vertical space.
 * Should be run after drag/resize ends to tidy up the grid.
 * 
 * @param {Array} allWidgets Array of widget objects
 * @returns {Array} New array of compacted widgets
 */
export function compactLayout(allWidgets) {
    const sorted = sortWidgets(allWidgets);
    const compacted = [];

    // Process each widget top-to-bottom
    for (const widget of sorted) {
        let newWidget = { ...widget };
        
        if (!newWidget.enabled) {
            compacted.push(newWidget);
            continue;
        }

        // Move widget up as far as possible
        // Start at y=0 and increment until no overlap with ALREADY PLACED widgets
        let y = 0;
        while (true) {
            const candidate = { ...newWidget, y };
            const collision = compacted.find(w => w.enabled && widgetsOverlap(candidate, w));
            
            if (!collision) {
                newWidget.y = y;
                break;
            }
            // Optimization: jump to bottom of colliding widget
            y = collision.y + collision.height;
        }
        
        compacted.push(newWidget);
    }

    return compacted;
}

/**
 * Finds the first available position for a new widget of given dimensions.
 * Scans top-to-bottom, left-to-right.
 * 
 * @param {Array} widgets Existing widgets
 * @param {number} width Width of new widget
 * @param {number} height Height of new widget
 * @param {number} gridCols Total grid columns (e.g. 16)
 * @returns {object} {x, y} coordinates
 */
export function findOpenPosition(widgets, width, height, gridCols = 16) {
    let y = 0;
    while (true) {
        for (let x = 0; x <= gridCols - width; x++) {
            const candidate = { x, y, width, height, id: '__candidate__' };
            const collision = widgets.find(w => w.enabled && widgetsOverlap(candidate, w));
            
            if (!collision) {
                return { x, y };
            }
        }
        y++;
        // Safety break
        if (y > 1000) return { x: 0, y: 0 }; // Should basically never happen
    }
}
