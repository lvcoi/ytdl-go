# Contributing to ytdl-go

Thanks for contributing to `ytdl-go`. This guide covers repository-wide contribution workflow.

## Development Workflow

1. Fork and clone the repository.
2. Create a branch:

   ```sh
   git checkout -b feature/your-change
   ```

3. Make your changes.
4. Run backend checks from repository root:

   ```sh
   go test ./...
   ```

5. Run frontend build checks:

   ```sh
   cd frontend
   npm run build
   ```

6. Optional integrated build check:

   ```sh
   ./build.sh
   ```

7. Commit, push, and open a pull request.

## Scope-Specific Contribution Guides

- Frontend-specific contribution details: `frontend/docs/CONTRIBUTING.md`
- Security reporting process: `SECURITY.md`

## Pull Request Expectations

- Keep PRs focused in scope.
- Include a clear summary of behavioral changes.
- Include validation steps you ran.
- Coordinate API contract changes across frontend and backend when applicable.
