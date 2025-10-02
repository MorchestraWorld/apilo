import { useState, useEffect, useRef } from 'react';
import { Search, X } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

interface SearchResult {
  title: string;
  path: string;
  description: string;
}

const searchablePages: SearchResult[] = [
  {
    title: 'Quick Start',
    path: '/docs/quickstart',
    description: '5-minute setup guide with Claude Code integration',
  },
  {
    title: 'Features',
    path: '/docs/features',
    description: 'Memory-bounded cache, circuit breaker, HTTP/2 optimization',
  },
  {
    title: 'Performance',
    path: '/docs/performance',
    description: '93.69% latency reduction, validated metrics, test methodology',
  },
  {
    title: 'Configuration',
    path: '/docs/configuration',
    description: 'YAML reference, environment variables, CLI flags',
  },
  {
    title: 'Claude Code Integration',
    path: '/integration/claude-code',
    description: 'Complete integration guide for Claude Code',
  },
];

export function SearchBar() {
  const [query, setQuery] = useState('');
  const [isOpen, setIsOpen] = useState(false);
  const [results, setResults] = useState<SearchResult[]>([]);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const searchRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();

  useEffect(() => {
    if (query.trim() === '') {
      setResults([]);
      setSelectedIndex(0);
      return;
    }

    const filtered = searchablePages.filter((page) => {
      const searchTerm = query.toLowerCase();
      return (
        page.title.toLowerCase().includes(searchTerm) ||
        page.description.toLowerCase().includes(searchTerm)
      );
    });

    setResults(filtered);
    setSelectedIndex(0);
  }, [query]);

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (searchRef.current && !searchRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      setIsOpen(false);
      setQuery('');
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      setSelectedIndex((prev) => (prev < results.length - 1 ? prev + 1 : prev));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setSelectedIndex((prev) => (prev > 0 ? prev - 1 : 0));
    } else if (e.key === 'Enter' && results.length > 0) {
      e.preventDefault();
      navigate(results[selectedIndex].path);
      setIsOpen(false);
      setQuery('');
    }
  };

  const handleSelect = (path: string) => {
    navigate(path);
    setIsOpen(false);
    setQuery('');
  };

  return (
    <div className="relative" ref={searchRef}>
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <input
          type="text"
          placeholder="Search docs..."
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
            setIsOpen(true);
          }}
          onFocus={() => setIsOpen(true)}
          onKeyDown={handleKeyDown}
          className="w-full md:w-64 pl-9 pr-9 py-2 text-sm bg-muted/50 border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
        />
        {query && (
          <button
            onClick={() => {
              setQuery('');
              setResults([]);
            }}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
          >
            <X className="h-4 w-4" />
          </button>
        )}
      </div>

      {/* Search Results Dropdown */}
      {isOpen && results.length > 0 && (
        <div className="absolute top-full mt-2 w-full md:w-96 bg-card border border-border rounded-lg shadow-lg overflow-hidden z-50">
          {results.map((result, index) => (
            <button
              key={result.path}
              onClick={() => handleSelect(result.path)}
              className={`w-full text-left px-4 py-3 hover:bg-muted/50 transition-colors border-b border-border last:border-b-0 ${
                index === selectedIndex ? 'bg-muted/50' : ''
              }`}
            >
              <div className="font-medium text-sm">{result.title}</div>
              <div className="text-xs text-muted-foreground mt-1">
                {result.description}
              </div>
            </button>
          ))}
        </div>
      )}

      {/* No Results */}
      {isOpen && query && results.length === 0 && (
        <div className="absolute top-full mt-2 w-full md:w-96 bg-card border border-border rounded-lg shadow-lg p-4 z-50">
          <p className="text-sm text-muted-foreground text-center">
            No results found for "{query}"
          </p>
        </div>
      )}
    </div>
  );
}
