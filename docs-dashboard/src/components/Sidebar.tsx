import { Link, useLocation } from 'react-router-dom';
import {
  Home,
  BookOpen,
  Zap,
  Settings,
  FileCode,
  Github,
  TrendingUp,
  Code,
  Menu,
  X
} from 'lucide-react';
import { useState } from 'react';

interface NavItem {
  path: string;
  label: string;
  icon: React.ElementType;
  children?: NavItem[];
}

const navigation: NavItem[] = [
  { path: '/', label: 'Home', icon: Home },
  {
    path: '/docs',
    label: 'Documentation',
    icon: BookOpen,
    children: [
      { path: '/docs/quickstart', label: 'Quick Start', icon: Zap },
      { path: '/docs/features', label: 'Features', icon: Settings },
      { path: '/docs/performance', label: 'Performance', icon: TrendingUp },
      { path: '/docs/configuration', label: 'Configuration', icon: FileCode },
    ]
  },
  {
    path: '/integration',
    label: 'Integration',
    icon: Code,
    children: [
      { path: '/integration/claude-code', label: 'Claude Code', icon: Code },
    ]
  },
];

export function Sidebar() {
  const location = useLocation();
  const [isOpen, setIsOpen] = useState(false);

  const isActive = (path: string) => {
    return location.pathname === path;
  };

  const NavLink = ({ item }: { item: NavItem }) => {
    const Icon = item.icon;
    const active = isActive(item.path);

    return (
      <Link
        to={item.path}
        className={active ? 'sidebar-link-active' : 'sidebar-link'}
        onClick={() => setIsOpen(false)}
      >
        <Icon className="w-4 h-4" />
        <span>{item.label}</span>
      </Link>
    );
  };

  return (
    <>
      {/* Mobile Menu Button */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="lg:hidden fixed top-4 left-4 z-50 p-2 rounded-lg bg-card border shadow-lg"
      >
        {isOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
      </button>

      {/* Overlay */}
      {isOpen && (
        <div
          className="lg:hidden fixed inset-0 bg-black/50 z-40"
          onClick={() => setIsOpen(false)}
        />
      )}

      {/* Sidebar */}
      <aside
        className={`
          fixed lg:sticky top-0 left-0 h-screen w-64 bg-card border-r
          transition-transform duration-300 z-40 flex flex-col
          ${isOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'}
        `}
      >
        {/* Logo */}
        <div className="p-6 border-b">
          <Link to="/" className="flex items-center gap-2" onClick={() => setIsOpen(false)}>
            <div className="gradient-bg p-2 rounded-lg">
              <Zap className="w-6 h-6 text-white" />
            </div>
            <div>
              <h1 className="font-bold text-lg">API Latency</h1>
              <p className="text-xs text-muted-foreground">Optimizer</p>
            </div>
          </Link>
        </div>

        {/* Navigation */}
        <nav className="flex-1 overflow-y-auto p-4 space-y-2">
          {navigation.map((item) => (
            <div key={item.path}>
              <NavLink item={item} />
              {item.children && (
                <div className="ml-4 mt-1 space-y-1">
                  {item.children.map((child) => (
                    <NavLink key={child.path} item={child} />
                  ))}
                </div>
              )}
            </div>
          ))}
        </nav>

        {/* Footer */}
        <div className="p-4 border-t">
          <a
            href="https://github.com/joshkornreich/api-latency-optimizer"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors"
          >
            <Github className="w-4 h-4" />
            <span>View on GitHub</span>
          </a>
        </div>
      </aside>
    </>
  );
}
