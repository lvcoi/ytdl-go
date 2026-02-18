import { For, Show } from 'solid-js';

export function Grid(props) {
    return (
        <div class={`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 ${props.class || ''}`}>
            {props.children}
        </div>
    );
}

export function GridItem(props) {
    const spanClass = () => {
        switch(props.span) {
            case 2: return 'lg:col-span-2';
            case 3: return 'lg:col-span-3';
            case 4: return 'lg:col-span-4';
            default: return 'lg:col-span-1';
        }
    };

    return (
        <div class={`${spanClass()} ${props.class || ''}`}>
            {props.children}
        </div>
    );
}
