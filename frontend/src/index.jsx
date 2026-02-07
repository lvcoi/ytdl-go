import { render } from 'solid-js/web';
import App from './App';
import { AppStoreProvider } from './store/appStore';
import '../index.css';

const root = document.getElementById('root');
render(() => (
  <AppStoreProvider>
    <App />
  </AppStoreProvider>
), root);
