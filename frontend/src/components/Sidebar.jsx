import { createSignal, onMount, onCleanup, Show } from 'solid-js';
import { A, useLocation } from '@solidjs/router';
import Icon from './Icon';
import logo from '../assets/logo.png';
import { useAppStore } from '../store/appStore';

export default function Sidebar() {
    const location = useLocation();
    const { state, setState } = useAppStore();

    const [isAccountDropdownOpen, setIsAccountDropdownOpen] = createSignal(false);
    const [windowWidth, setWindowWidth] = createSignal(
        typeof window !== 'undefined' ? window.innerWidth : 1024
    );

    onMount(() => {
        const handleResize = () => setWindowWidth(window.innerWidth);
        window.addEventListener('resize', handleResize);
        onCleanup(() => window.removeEventListener('resize', handleResize));
    });

    const isEffectivelyCollapsed = () => windowWidth() < 768 || state.ui.isSidebarCollapsed;
    const isSettings = () => location.pathname.startsWith('/settings');

    const toggleSidebar = () => {
        setState('ui', 'isSidebarCollapsed', (prev) => !prev);
        setIsAccountDropdownOpen(false);
    };

    const switchAccount = (accountName) => {
        setState('ui', 'activeAccount', accountName);
        setIsAccountDropdownOpen(false);
    };

    const active = (path) => {
        if (path === '/settings') {
            return isSettings()
                ? 'bg-accent-primary/10 text-accent-secondary shadow-sm shadow-accent-primary/5'
                : 'text-gray-500 hover:bg-white/5 hover:text-gray-300';
        }
        return location.pathname === path
            ? 'bg-accent-primary/10 text-accent-secondary shadow-sm shadow-accent-primary/5'
            : 'text-gray-500 hover:bg-white/5 hover:text-gray-300';
    };

    const NavLink = (props) => (
        <div class="relative group/nav">
            <A
                href={props.href}
                class={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-200 ${active(props.href)} ${isEffectivelyCollapsed() ? 'justify-center px-0' : ''}`}
            >
                <Icon name={props.icon} class="w-5 h-5 shrink-0" />
                <Show when={!isEffectivelyCollapsed()}>
                    <span class="font-bold text-sm">{props.label}</span>
                </Show>
            </A>
            <Show when={isEffectivelyCollapsed()}>
                <span class="pointer-events-none absolute left-full top-1/2 -translate-y-1/2 ml-3 px-2 py-1 rounded-lg bg-bg-surface-soft border border-white/10 text-xs font-bold text-gray-200 whitespace-nowrap opacity-0 group-hover/nav:opacity-100 transition-opacity duration-150 shadow-xl z-50">
                    {props.label}
                </span>
            </Show>
        </div>
    );

    const SettingsNavLink = (props) => (
        <div class="relative group/nav">
            <A
                href={props.href}
                class={`w-full flex items-center gap-3 px-4 py-2.5 rounded-lg transition-all duration-200 text-sm ${
                    location.pathname === props.href
                        ? 'bg-white/10 text-white font-bold'
                        : 'text-gray-400 hover:bg-white/5 hover:text-gray-200 font-medium'
                } ${isEffectivelyCollapsed() ? 'justify-center px-0' : ''}`}
            >
                <Icon name={props.icon} class="w-4 h-4 shrink-0" />
                <Show when={!isEffectivelyCollapsed()}>
                    <span>{props.label}</span>
                </Show>
            </A>
            <Show when={isEffectivelyCollapsed()}>
                <span class="pointer-events-none absolute left-full top-1/2 -translate-y-1/2 ml-3 px-2 py-1 rounded-lg bg-bg-surface-soft border border-white/10 text-xs font-bold text-gray-200 whitespace-nowrap opacity-0 group-hover/nav:opacity-100 transition-opacity duration-150 shadow-xl z-50">
                    {props.label}
                </span>
            </Show>
        </div>
    );

    return (
        <aside class={`${isEffectivelyCollapsed() ? 'w-16' : 'w-72'} max-md:w-16 bg-bg-surface/85 border-r border-white/10 flex flex-col ${isEffectivelyCollapsed() ? 'p-3' : 'p-6'} backdrop-blur-xl transition-all duration-300 shrink-0`}>
            <div class={`flex items-center ${isEffectivelyCollapsed() ? 'justify-center' : 'gap-3'} mb-10 ${isEffectivelyCollapsed() ? '' : 'px-2'}`}>
                <div class="w-10 h-10 bg-accent-primary/10 rounded-2xl flex items-center justify-center shadow-lg shadow-accent-primary/5 overflow-hidden border border-accent-primary/20 shrink-0">
                    <img src={logo} alt="ytdl-go logo" class="w-8 h-8 object-contain brightness-0 invert opacity-80" />
                </div>
                <Show when={!isEffectivelyCollapsed()}>
                    <span class="text-xl font-black tracking-tight text-transparent bg-clip-text bg-vibrant-gradient">ytdl-go</span>
                </Show>
            </div>

            {/* Account Switcher */}
            <Show when={!isEffectivelyCollapsed()}>
                <div class="mb-6 relative">
                    <button
                        type="button"
                        onClick={() => setIsAccountDropdownOpen((prev) => !prev)}
                        class="w-full flex items-center justify-between gap-2 px-4 py-2.5 rounded-xl border border-white/10 bg-white/5 hover:bg-white/10 transition-colors text-sm font-bold text-gray-200"
                    >
                        <div class="flex items-center gap-2 min-w-0">
                            <div class="w-6 h-6 rounded-full bg-accent-primary/20 flex items-center justify-center text-[10px] font-black text-accent-primary shrink-0">
                                {state.ui.activeAccount.charAt(0).toUpperCase()}
                            </div>
                            <span class="truncate">{state.ui.activeAccount}</span>
                        </div>
                        <Icon name="chevrons-up-down" class="w-4 h-4 text-gray-500 shrink-0" />
                    </button>

                    <Show when={isAccountDropdownOpen()}>
                        <div class="absolute left-0 right-0 top-full mt-1 z-50 rounded-xl border border-white/10 bg-gray-900/95 backdrop-blur-xl shadow-2xl overflow-hidden">
                            <div class="p-1.5">
                                {['Personal', 'Work'].map((accountName) => (
                                    <button
                                        type="button"
                                        onClick={() => switchAccount(accountName)}
                                        class={`w-full flex items-center gap-2.5 px-3 py-2 rounded-lg text-left text-sm transition-colors duration-100 ${
                                            state.ui.activeAccount === accountName
                                                ? 'bg-accent-primary/10 text-white font-bold'
                                                : 'text-gray-300 hover:bg-white/10 hover:text-white font-medium'
                                        }`}
                                    >
                                        <div class="w-5 h-5 rounded-full bg-accent-primary/20 flex items-center justify-center text-[9px] font-black text-accent-primary shrink-0">
                                            {accountName.charAt(0).toUpperCase()}
                                        </div>
                                        {accountName}
                                    </button>
                                ))}
                            </div>
                        </div>
                    </Show>
                </div>
            </Show>

            <nav class="flex-1 space-y-2">
                <NavLink href="/" icon="layout-dashboard" label="Dashboard" />

                <Show when={!isEffectivelyCollapsed()}>
                    <div class="pt-4 pb-2 px-4 text-xs font-bold text-gray-600 uppercase tracking-widest">
                        Media
                    </div>
                </Show>
                <Show when={isEffectivelyCollapsed()}>
                    <div class="pt-2 pb-1 border-t border-white/5" />
                </Show>

                <NavLink href="/download" icon="plus-circle" label="New Download" />
                <NavLink href="/library" icon="layers" label="Library" />

                <Show when={!isEffectivelyCollapsed()}>
                    <div class="pt-4 pb-2 px-4 text-xs font-bold text-gray-600 uppercase tracking-widest">
                        System
                    </div>
                </Show>
                <Show when={isEffectivelyCollapsed()}>
                    <div class="pt-2 pb-1 border-t border-white/5" />
                </Show>

                <NavLink href="/settings" icon="sliders" label="Settings" />

                {/* Contextual Settings Sub-Navigation */}
                <Show when={isSettings()}>
                    <div class={`space-y-1 ${isEffectivelyCollapsed() ? '' : 'ml-4 pl-3 border-l border-white/10'}`}>
                        <SettingsNavLink href="/settings" icon="settings" label="General" />
                        <SettingsNavLink href="/settings/network" icon="wifi" label="Network" />
                    </div>
                </Show>
            </nav>

            <div class="mt-auto space-y-3">
                <Show when={!isEffectivelyCollapsed()}>
                    <div class="relative group/ext">
                        <span class="pointer-events-none absolute bottom-full left-1/2 -translate-x-1/2 mb-2 px-3 py-2 rounded-lg bg-bg-surface-soft border border-white/10 text-[10px] text-gray-300 w-56 text-center leading-relaxed opacity-0 group-hover/ext:opacity-100 transition-opacity duration-150 shadow-xl z-50 whitespace-normal">
                            Coming Soon: extension provider management and health status.
                        </span>
                        <button
                            type="button"
                            disabled
                            aria-disabled="true"
                            aria-label="Extensions panel (Coming Soon)"
                            class="w-full p-4 bg-white/5 rounded-2xl border border-white/5 opacity-70 cursor-not-allowed text-left hover:bg-white/10 transition-colors"
                        >
                            <div class="flex items-center gap-2 mb-2 text-xs font-bold text-gray-500 uppercase tracking-widest">
                                <Icon name="puzzle" class="w-3 h-3" />
                                Extensions
                            </div>
                            <div class="flex items-center justify-between text-xs">
                                <span class="text-gray-500">PO Token Provider</span>
                                <span class="px-2 py-0.5 bg-white/10 text-gray-500 rounded-full font-bold border border-white/10">
                                    Coming Soon
                                </span>
                            </div>
                        </button>
                    </div>
                </Show>
                <Show when={isEffectivelyCollapsed()}>
                    <div class="relative group/ext flex justify-center">
                        <button
                            type="button"
                            disabled
                            aria-disabled="true"
                            class="p-3 bg-white/5 rounded-xl border border-white/5 opacity-70 cursor-not-allowed hover:bg-white/10 transition-colors"
                        >
                            <Icon name="puzzle" class="w-5 h-5 text-gray-500" />
                        </button>
                        <span class="pointer-events-none absolute left-full top-1/2 -translate-y-1/2 ml-3 px-2 py-1 rounded-lg bg-bg-surface-soft border border-white/10 text-xs font-bold text-gray-200 whitespace-nowrap opacity-0 group-hover/ext:opacity-100 transition-opacity duration-150 shadow-xl z-50">
                            Extensions (Coming Soon)
                        </span>
                    </div>
                </Show>

                <button
                    type="button"
                    onClick={toggleSidebar}
                    aria-label={state.ui.isSidebarCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
                    class={`w-full flex items-center ${isEffectivelyCollapsed() ? 'justify-center px-0 py-3' : 'gap-2 px-4 py-2'} rounded-xl text-gray-600 hover:text-gray-400 hover:bg-white/5 transition-all duration-200 text-xs font-bold`}
                >
                    <Icon name={state.ui.isSidebarCollapsed ? 'panel-left-open' : 'panel-left-close'} class="w-4 h-4 transition-transform duration-300" />
                    <Show when={!isEffectivelyCollapsed()}>
                        <span>Collapse</span>
                    </Show>
                </button>
            </div>
        </aside>
    );
}
