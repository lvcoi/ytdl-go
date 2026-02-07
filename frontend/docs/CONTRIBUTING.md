# Contributing to the Frontend

We welcome contributions! Whether you're fixing a bug, improving the design, or adding a new feature, here's how to get started.

## Development Workflow

1. **Fork & Clone:** Fork the repository and clone it locally.
2. **Install:** Run `npm install` inside the `frontend/` folder.
3. **Branch:** Create a feature branch: `git checkout -b feature/amazing-new-view`.
4. **Code:** Make your changes.
   - **Hot Reloading:** Use `npm run dev` to see changes instantly.
5. **Test:** Ensure everything works in the browser console (no red errors!).
6. **Build:** Run `npm run build` to ensure the build passes and files are generated correctly.
   - **Optional full-stack check:** From the repo root, run `./build.sh` to validate integrated backend + frontend build behavior.
7. **Push & PR:** Push your branch and open a Pull Request.

## Code Style Guidelines

- **Components:** Use functional components. Keep them small and focused.
- **File Naming:** PascalCase for components (e.g., `DownloadView.jsx`), camelCase for logic/utils.
- **Styling:** Prefer **Tailwind CSS** utility classes over custom CSS. Only use `index.css` for highly specific global overrides (like scrollbars).
- **Icons:** Use `lucide` icons via `createIcons()`. Always ensure icons are re-initialized in `createEffect` if the DOM structure changes significantly.

## Adding New Dependencies

Please discuss adding large new dependencies in an Issue first. We aim to keep the bundle size small to ensure the Go binary remains lightweight.

- **Allowed:** Small utility libraries, UI helpers.
- **Avoid:** Heavy component libraries (MUI, Bootstrap), massive data visualization libraries (unless properly tree-shaken).
