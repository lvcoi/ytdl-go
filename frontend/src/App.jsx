import { Router, Route, Navigate } from '@solidjs/router';
import { lazy, onMount } from 'solid-js';
import MainLayout from './layouts/MainLayout';
import { useSavedPlaylists } from './hooks/useSavedPlaylists';
import { useLibrarySync } from './hooks/useLibrarySync';
import { useDownloadManager } from './hooks/useDownloadManager';
import wsService from './services/websocket';

// Route Components
import { Dashboard } from './routes/Dashboard';
import { Download } from './routes/Download';
import { Library } from './routes/Library';
import { Settings } from './routes/Settings';

// Lazy loaded contextual routes
const NetworkSettings = lazy(() => import('./routes/NetworkSettings'));

function App() {
  const { initialize: initializeSavedPlaylists } = useSavedPlaylists();
  useLibrarySync();
  const { listenForProgress } = useDownloadManager();

  onMount(() => {
    void initializeSavedPlaylists();
    wsService.connect();
  });

  return (
    <Router root={MainLayout}>
      <Route path="/" component={Dashboard} />
      <Route path="/download" component={Download} />
      <Route path="/library" component={Library} />

      {/* Nested Settings Routing Context */}
      <Route path="/settings">
        <Route path="/" component={Settings} />
        <Route path="/network" component={NetworkSettings} />
        <Route path="/*" component={() => <Navigate href="/settings" />} />
      </Route>
    </Router>
  );
}

export default App;
