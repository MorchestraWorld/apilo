import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { lazy, Suspense } from 'react';
import { Sidebar } from './components/Sidebar';
import { ThemeToggle } from './components/ThemeToggle';
import { SearchBar } from './components/SearchBar';

// Lazy load page components
const Home = lazy(() => import('./pages/Home').then(m => ({ default: m.Home })));
const Docs = lazy(() => import('./pages/Docs').then(m => ({ default: m.Docs })));

// Loading fallback component
const PageLoader = () => (
  <div className="flex items-center justify-center min-h-[400px]">
    <div className="text-center">
      <div className="inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-primary border-r-transparent"></div>
      <p className="mt-4 text-muted-foreground">Loading...</p>
    </div>
  </div>
);

function App() {
  return (
    <Router>
      <div className="flex min-h-screen">
        {/* Sidebar */}
        <Sidebar />

        {/* Main Content */}
        <div className="flex-1 flex flex-col lg:ml-0">
          {/* Header */}
          <header className="sticky top-0 z-30 border-b bg-card/95 backdrop-blur supports-[backdrop-filter]:bg-card/60">
            <div className="container flex h-16 items-center justify-between gap-4 px-4 md:px-8">
              <h2 className="text-lg font-semibold hidden md:block whitespace-nowrap">API Latency Optimizer Docs</h2>
              <div className="flex-1 flex items-center justify-center md:justify-end gap-4">
                <SearchBar />
                <ThemeToggle />
              </div>
            </div>
          </header>

          {/* Page Content */}
          <main className="flex-1 p-4 md:p-8">
            <Suspense fallback={<PageLoader />}>
              <Routes>
              <Route path="/" element={<Home />} />
              <Route
                path="/docs/quickstart"
                element={
                  <Docs
                    contentPath="/content/quickstart.md"
                    title="Quick Start"
                  />
                }
              />
              <Route
                path="/docs/features"
                element={
                  <Docs
                    contentPath="/content/features.md"
                    title="Features"
                  />
                }
              />
              <Route
                path="/docs/performance"
                element={
                  <Docs
                    contentPath="/content/performance.md"
                    title="Performance"
                  />
                }
              />
              <Route
                path="/docs/configuration"
                element={
                  <Docs
                    contentPath="/content/configuration.md"
                    title="Configuration"
                  />
                }
              />
              <Route
                path="/integration/claude-code"
                element={
                  <Docs
                    contentPath="/content/integration.md"
                    title="Claude Code Integration"
                  />
                }
              />
              <Route path="/docs" element={<Home />} />
              <Route path="/integration" element={<Home />} />
            </Routes>
            </Suspense>
          </main>

          {/* Footer */}
          <footer className="border-t py-6 px-4 md:px-8">
            <div className="container mx-auto text-center text-sm text-muted-foreground">
              <p>
                Built with React, TypeScript, and Tailwind CSS.{' '}
                <a
                  href="https://github.com/joshkornreich/api-latency-optimizer"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="underline underline-offset-4 hover:text-foreground"
                >
                  View on GitHub
                </a>
              </p>
            </div>
          </footer>
        </div>
      </div>
    </Router>
  );
}

export default App;
