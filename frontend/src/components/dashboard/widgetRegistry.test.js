import { describe, it, expect } from 'vitest';
import { WIDGET_REGISTRY, DEFAULT_LAYOUT_WIDGETS, GRID_COLS } from './widgetRegistry';

describe('widgetRegistry', () => {
    it('defines valid grid constants', () => {
        expect(GRID_COLS).toBe(16);
    });

    it('contains all required widgets with valid metadata', () => {
        const requiredWidgets = ['welcome', 'quick-download', 'active-downloads', 'recent-activity', 'stats', 'concurrency'];
        
        requiredWidgets.forEach(id => {
            const widget = WIDGET_REGISTRY[id];
            expect(widget).toBeDefined();
            expect(widget.id).toBe(id);
            expect(widget.label).toBeTruthy();
            expect(widget.icon).toBeTruthy();
            expect(widget.defaultW).toBeGreaterThan(0);
            expect(widget.defaultH).toBeGreaterThan(0);
            expect(widget.minW).toBeGreaterThan(0);
            expect(widget.minH).toBeGreaterThan(0);
            
            // Validation: Min size should not exceed default size
            expect(widget.minW).toBeLessThanOrEqual(widget.defaultW);
            expect(widget.minH).toBeLessThanOrEqual(widget.defaultH);
        });
    });

    it('defines a valid default layout', () => {
        expect(Array.isArray(DEFAULT_LAYOUT_WIDGETS)).toBe(true);
        expect(DEFAULT_LAYOUT_WIDGETS.length).toBeGreaterThan(0);
        
        DEFAULT_LAYOUT_WIDGETS.forEach(widget => {
            expect(widget.id).toBeTruthy();
            expect(typeof widget.x).toBe('number');
            expect(typeof widget.y).toBe('number');
            expect(typeof widget.width).toBe('number');
            expect(typeof widget.height).toBe('number');
            
            // Bounds check
            expect(widget.x).toBeGreaterThanOrEqual(0);
            expect(widget.x + widget.width).toBeLessThanOrEqual(GRID_COLS);
        });
        
        // Ensure ConcurrencyWidget is present but disabled by default
        const concurrency = DEFAULT_LAYOUT_WIDGETS.find(w => w.id === 'concurrency');
        expect(concurrency).toBeDefined();
        expect(concurrency.enabled).toBe(false);
    });
});
