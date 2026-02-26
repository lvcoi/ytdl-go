import { createSignal, Show } from 'solid-js';

export default function Tooltip(props) {
    const [visible, setVisible] = createSignal(false);
    let showTimeout;

    const onEnter = () => {
        showTimeout = setTimeout(() => setVisible(true), props.delay ?? 200);
    };

    const onLeave = () => {
        clearTimeout(showTimeout);
        setVisible(false);
    };

    const positionClasses = () => {
        const pos = props.position ?? 'top';
        if (pos === 'bottom') return 'top-full mt-2 left-1/2 -translate-x-1/2';
        return 'bottom-full mb-2 left-1/2 -translate-x-1/2';
    };

    return (
        <div
            class="relative inline-flex"
            onMouseEnter={onEnter}
            onMouseLeave={onLeave}
            onFocusIn={onEnter}
            onFocusOut={onLeave}
        >
            {props.children}

            <Show when={visible()}>
                <div
                    class={`absolute z-50 pointer-events-none whitespace-nowrap
                            px-2.5 py-1.5 rounded-lg text-xs font-semibold
                            bg-gray-900/95 text-gray-100 border border-white/10
                            shadow-xl backdrop-blur-sm
                            animate-in fade-in zoom-in-95 duration-150
                            ${positionClasses()}`}
                    role="tooltip"
                >
                    {props.text}
                </div>
            </Show>
        </div>
    );
}
