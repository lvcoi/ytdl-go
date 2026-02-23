import { render } from '@solidjs/testing-library';
import { MemoryRouter, Route } from '@solidjs/router';
import { AppStoreProvider } from '../store/appStore';

/**
 * Renders a component inside a MemoryRouter + AppStoreProvider context.
 * Use this for components that depend on @solidjs/router primitives
 * (useLocation, useNavigate, useSearchParams, <A>, etc.)
 * and/or useAppStore().
 *
 * @param {Function} Component - JSX factory: () => <MyComponent />
 * @param {object} [options] - Optional overrides
 * @param {string} [options.initialPath='/'] - Starting URL path
 * @returns render result from @solidjs/testing-library
 */
export function renderWithRouter(Component, { initialPath = '/' } = {}) {
    return render(() => (
        <AppStoreProvider>
            <MemoryRouter initialEntries={[initialPath]}>
                <Route path="*" component={() => Component()} />
            </MemoryRouter>
        </AppStoreProvider>
    ));
}
