# ytdl-go Web UI

The modern, reactive web interface for [ytdl-go](https://github.com/lvcoi/ytdl-go). Built with performance and user experience in mind, this frontend provides a sleek dashboard for managing downloads, viewing your library, and configuring application settings.

## 🚀 Tech Stack

- **Framework:** [SolidJS](https://www.solidjs.com/) - High-performance reactive UI library.
- **Build Tool:** [Vite](https://vitejs.dev/) - Next-generation frontend tooling.
- **Styling:** [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS framework.
- **Icons:** [Lucide](https://lucide.dev/) - Beautiful & consistent open-source icons.

## 🛠️ Prerequisites

- **Node.js:** Version 18.0.0 or higher.
- **npm:** Installed automatically with Node.js.

## 📦 Installation

Navigate to the frontend directory and install dependencies:

```sh
cd frontend
npm install
```

## 💻 Development

Start the local development server with hot module replacement (HMR):

```sh
npm run dev
```

The app will be available at `http://localhost:5173` (or similar).

> **Note:** To enable API calls in development mode, configure the proxy in `vite.config.js`:
>
> ```js
> server: {
>   proxy: {
>     '/api': 'http://localhost:8080'
>   }
> }
> ```

## 🏗️ Building for Production

To compile the frontend for integration with the Go backend:

```sh
npm run build
```

**Output Behavior:**
This command compiles the source code and outputs optimized static files (`index.html`, `app.js`, `styles.css`) directly into the Go project's asset directory:
`../internal/web/assets/`

The Go server is configured to serve these files automatically.

## 📂 Project Structure

```text
frontend/
├── src/
│   ├── components/
│   │   ├── DownloadView.jsx  # Download input & status
│   │   ├── LibraryView.jsx   # Media library grid
│   │   ├── Player.jsx        # Floating media player
│   │   └── SettingsView.jsx  # Configuration panel
│   ├── App.jsx               # Main layout & routing state
│   ├── index.jsx             # App entry point
│   └── index.css             # Tailwind imports & global styles
├── index.html                # Vite entry template
├── package.json              # Dependencies & scripts
├── postcss.config.js         # PostCSS configuration
├── tailwind.config.js        # Tailwind configuration
└── vite.config.js            # Vite build & proxy settings
```

## 📖 Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [API Reference](docs/API.md)
- [Contributing](docs/CONTRIBUTING.md)
