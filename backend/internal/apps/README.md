# internal/apps

`internal/apps` is the application boundary layer for modular-monolith migration.

Current split:

- `shared`: cross-domain routes and shared modules
- `attendance`: attendance domain handler/service/repository modules
- `attendance_routes`: attendance domain route registration
- `expense`: expense domain handler/service/repository modules
- `expense_routes`: expense domain route registration
- `hr`: HR domain handler/service/repository modules
- `hr_routes`: HR domain route registration

Migration rule:

- Add new feature routes under each app package first.
- Keep `internal/router` as orchestration only.
- Gradually move handler/service/repository from monolithic packages to each app package.
