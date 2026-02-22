# Contributing to the Frontend

This guide covers contributions to the web UI under `frontend/`.
For repository-wide contribution workflow, see `docs/CONTRIBUTING.md`.

## Frontend Workflow

1. Fork and clone the repository.
2. Install frontend dependencies:

   ```sh
   cd frontend
   npm install
   ```

3. Create a feature branch:

   ```sh
   git checkout -b feature/frontend-change
   ```

4. Implement your change in `frontend/src/`.
5. Validate the frontend build:

   ```sh
   npm run build
   ```

6. If your change touches API behavior or integration paths, run backend tests from repo root:

   ```sh
   go test ./...
   ```

7. Push and open a pull request.

## Frontend Conventions

- Components: keep components focused and colocate related logic where practical.
- Naming: PascalCase for component files, camelCase for helpers/utilities.
- Styling: prefer Tailwind utility classes, keeping global CSS in `frontend/index.css` minimal.
- Icons: use `lucide-solid` through `frontend/src/components/Icon.jsx`; add icons by importing them there and extending `iconMap`.

## Dependency Policy

Keep dependencies lean.

- Preferred: small focused libraries.
- Avoid by default: large UI frameworks and heavy libraries that increase bundle size significantly.
