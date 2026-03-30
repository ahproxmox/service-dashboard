# Service Icons

Minimalist SVG icons for Service Dashboard services.

## Icon Naming
- Each icon is named after the service (e.g., kanban.svg, rag.svg)
- Icons use `currentColor` for stroke, allowing color theming
- Icons are 100x100 viewBox for consistent sizing
- Color is applied via the service configuration in services-config.json

## Adding New Icons
1. Create new SVG file with service name
2. Add to frontend/public/icons/
3. Update services-config.json with icon path
4. Icon will automatically be used in ServiceCard component
