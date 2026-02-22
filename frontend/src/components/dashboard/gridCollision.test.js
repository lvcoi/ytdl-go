import { describe, it, expect } from 'vitest';
import { 
    widgetsOverlap, 
    sortWidgets, 
    resolveCollisions, 
    compactLayout, 
    findOpenPosition 
} from './gridCollision';

describe('gridCollision', () => {
    describe('widgetsOverlap', () => {
        it('detects overlap correctly', () => {
            const a = { x: 0, y: 0, width: 4, height: 2, id: 'a' };
            const b = { x: 2, y: 1, width: 4, height: 2, id: 'b' }; // Overlaps
            const c = { x: 4, y: 0, width: 4, height: 2, id: 'c' }; // Adjacent right
            const d = { x: 0, y: 2, width: 4, height: 2, id: 'd' }; // Adjacent bottom
            
            expect(widgetsOverlap(a, b)).toBe(true);
            expect(widgetsOverlap(a, c)).toBe(false);
            expect(widgetsOverlap(a, d)).toBe(false);
        });
    });

    describe('sortWidgets', () => {
        it('sorts top-to-bottom, left-to-right', () => {
            const widgets = [
                { id: 'c', x: 0, y: 2 },
                { id: 'b', x: 4, y: 0 },
                { id: 'a', x: 0, y: 0 }
            ];
            
            const sorted = sortWidgets(widgets);
            expect(sorted.map(w => w.id)).toEqual(['a', 'b', 'c']);
        });
    });

    describe('resolveCollisions', () => {
        it('pushes overlapping widgets downward', () => {
            const widgets = [
                { id: 'a', x: 0, y: 0, width: 4, height: 2, enabled: true }, // Moving widget
                { id: 'b', x: 0, y: 1, width: 4, height: 2, enabled: true }  // Target
            ];
            
            // Move 'a' down to y=1, causing overlap with 'b'
            const movedA = { ...widgets[0], y: 1 };
            
            const resolved = resolveCollisions(widgets, movedA);
            const resB = resolved.find(w => w.id === 'b');
            
            // 'b' should be pushed to y = movedA.y + movedA.height = 1 + 2 = 3
            expect(resB.y).toBe(3);
        });

        it('cascades push to subsequent widgets', () => {
            const widgets = [
                { id: 'a', x: 0, y: 0, width: 4, height: 2, enabled: true },
                { id: 'b', x: 0, y: 2, width: 4, height: 2, enabled: true },
                { id: 'c', x: 0, y: 4, width: 4, height: 2, enabled: true }
            ];
            
            // Move 'a' to y=1. Overlaps 'b' (at 2) partially? 
            // a (1-3). b (2-4). Overlap!
            const movedA = { ...widgets[0], y: 1 };
            
            const resolved = resolveCollisions(widgets, movedA);
            const resB = resolved.find(w => w.id === 'b');
            const resC = resolved.find(w => w.id === 'c');
            
            // b pushed to 1+2 = 3
            expect(resB.y).toBe(3);
            
            // b (3-5) overlaps c (4-6).
            // c pushed to 3+2 = 5
            expect(resC.y).toBe(5);
        });
    });

    describe('compactLayout', () => {
        it('moves widgets up to fill gaps', () => {
            const widgets = [
                { id: 'a', x: 0, y: 0, width: 4, height: 2, enabled: true },
                { id: 'b', x: 0, y: 5, width: 4, height: 2, enabled: true } // Gap between 2 and 5
            ];
            
            const compacted = compactLayout(widgets);
            const resB = compacted.find(w => w.id === 'b');
            
            // b should move to y=2
            expect(resB.y).toBe(2);
        });
    });

    describe('findOpenPosition', () => {
        it('finds first available spot', () => {
            const widgets = [
                { id: 'a', x: 0, y: 0, width: 4, height: 2, enabled: true }
            ];
            
            // Request 4x2 widget
            const pos = findOpenPosition(widgets, 4, 2, 16);
            
            // 0,0 is taken. 4,0 is open.
            expect(pos).toEqual({ x: 4, y: 0 });
        });
    });
});
