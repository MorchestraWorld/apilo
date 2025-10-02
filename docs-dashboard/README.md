# API Latency Optimizer - Documentation Dashboard

A production-ready React documentation dashboard showcasing the **93.69% latency reduction** achieved by the API Latency Optimizer.

## Features

- **Modern Tech Stack**: React 19, TypeScript, Vite 7, Tailwind CSS 4
- **Code Splitting**: Lazy-loaded routes reduce initial bundle from 1.26 MB to 240 KB (80% reduction)
- **Dark/Light Mode**: Persistent theme with system preference detection
- **Responsive Design**: Mobile-first approach with sidebar navigation
- **Markdown Documentation**: GitHub Flavored Markdown with syntax highlighting
- **Performance Optimized**: Fast builds, HMR, and production-ready output

## Quick Start

```bash
# Install dependencies
npm install

# Development mode (with hot reload)
npm run dev

# Production build
npm run build

# Preview production build
npm run preview
```

## Project Structure

```
docs-dashboard/
├── public/content/          # Markdown documentation files
│   ├── quickstart.md
│   ├── features.md
│   ├── performance.md
│   ├── configuration.md
│   └── integration.md
├── src/
│   ├── components/          # Reusable UI components
│   │   ├── Sidebar.tsx
│   │   ├── ThemeToggle.tsx
│   │   ├── MarkdownViewer.tsx
│   │   ├── CodeBlock.tsx
│   │   ├── MetricsCard.tsx
│   │   └── FeatureCard.tsx
│   ├── pages/               # Page components (lazy-loaded)
│   │   ├── Home.tsx
│   │   └── Docs.tsx
│   ├── hooks/
│   │   └── useTheme.ts
│   ├── App.tsx              # Main app with routing
│   ├── main.tsx
│   └── index.css
└── dist/                    # Production build output
```

## Bundle Size (After Optimization)

- **Initial Load**: 240.22 kB (gzip: 76.47 kB) - 80% reduction
- **Home Page**: 10.60 kB (gzip: 2.84 kB) - Lazy loaded
- **Docs Pages**: 1,009.99 kB (gzip: 336.73 kB) - Lazy loaded
- **CSS**: 22.09 kB (gzip: 5.00 kB)

## Key Performance Metrics Displayed

- **93.69%** latency reduction
- **15.8x** throughput increase
- **98%** cache hit ratio

## Adding New Documentation

1. Create markdown file in `public/content/`
2. Add route in `src/App.tsx`:
   ```tsx
   <Route path="/docs/your-page" element={
     <Docs contentPath="/content/your-page.md" title="Your Page" />
   } />
   ```
3. Add navigation link in `src/components/Sidebar.tsx`

## Customization

### Update Theme Colors
Edit `src/index.css`:
```css
:root {
  --primary-500: #667eea;    /* Purple gradient start */
  --secondary-500: #764ba2;  /* Purple gradient end */
}
```

### Modify Branding
Update logo and title in `src/components/Sidebar.tsx`

## Development

- **TypeScript**: Full type safety with strict mode
- **ESLint**: Code quality and consistency
- **Hot Module Replacement**: Instant updates during development
- **Vite**: Lightning-fast builds and dev server

## Deployment

Build for production and deploy the `dist/` directory to:
- Vercel
- Netlify
- GitHub Pages
- Any static hosting service

```bash
npm run build
# Deploy the dist/ directory
```

## License

Part of the API Latency Optimizer project.
