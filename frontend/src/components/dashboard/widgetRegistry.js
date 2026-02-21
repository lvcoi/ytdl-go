// Widget content-based minimum sizes and metadata
// Based on 80px row height (~75px per column unit)
export const WIDGET_REGISTRY = {
    'welcome': {
        id: 'welcome',
        label: 'Welcome',
        icon: 'home',
        defaultW: 16,
        defaultH: 3,
        minW: 6,
        minH: 3
    },
    'quick-download': {
        id: 'quick-download',
        label: 'Quick Download',
        icon: 'download-cloud',
        defaultW: 8,
        defaultH: 2,
        minW: 6,
        minH: 2
    },
    'active-downloads': {
        id: 'active-downloads',
        label: 'Active Downloads',
        icon: 'download-cloud',
        defaultW: 4,
        defaultH: 4,
        minW: 4,
        minH: 3
    },
    'recent-activity': {
        id: 'recent-activity',
        label: 'Recent Activity',
        icon: 'history',
        defaultW: 12,
        defaultH: 4,
        minW: 6,
        minH: 4
    },
    'stats': {
        id: 'stats',
        label: 'Library Stats',
        icon: 'bar-chart-2',
        defaultW: 4,
        defaultH: 3,
        minW: 4,
        minH: 3
    },
    'concurrency': {
        id: 'concurrency',
        label: 'Concurrency',
        icon: 'settings',
        defaultW: 4,
        defaultH: 4,
        minW: 4,
        minH: 2
    },
};

// Default layout using 16-col coordinates
// Migration from v2 (4-col): x*4, width*4
export const DEFAULT_LAYOUT_WIDGETS = [
    { id: 'welcome', enabled: true, x: 0, y: 0, width: 16, height: 3 },
    { id: 'quick-download', enabled: true, x: 0, y: 3, width: 8, height: 2 },
    { id: 'active-downloads', enabled: true, x: 12, y: 3, width: 4, height: 4 },
    { id: 'recent-activity', enabled: true, x: 0, y: 5, width: 12, height: 4 },
    { id: 'stats', enabled: true, x: 12, y: 7, width: 4, height: 3 },
    { id: 'concurrency', enabled: false, x: 8, y: 3, width: 4, height: 2 }, // Default disabled, placed in gap
];

export const GRID_COLS = 16;
export const GRID_ROW_HEIGHT_PX = 80;
export const GRID_GAP_PX = 12;
