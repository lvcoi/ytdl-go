import { Router, Route } from '@solidjs/router';
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

function App() {
  const { initialize: initializeSavedPlaylists } = useSavedPlaylists();
  useLibrarySync();
  const { listenForProgress } = useDownloadManager(); // Ensure listeners are active

  onMount(() => {
    void initializeSavedPlaylists();
    wsService.connect();
  });

  return (
    <Router root={MainLayout}>
      <Route path="/" component={Dashboard} />
      <Route path="/download" component={Download} />
      <Route path="/library" component={Library} />
      <Route path="/settings" component={Settings} />
    </Router>
  );
}

export default App;
