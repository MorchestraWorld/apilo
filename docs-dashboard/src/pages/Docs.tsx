import { useEffect, useState } from 'react';
import { MarkdownViewer } from '../components/MarkdownViewer';
import { Loader2 } from 'lucide-react';

interface DocsProps {
  contentPath: string;
  title: string;
}

export function Docs({ contentPath, title }: DocsProps) {
  const [content, setContent] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadContent = async () => {
      try {
        setLoading(true);
        setError(null);
        const response = await fetch(contentPath);
        if (!response.ok) {
          throw new Error(`Failed to load content: ${response.statusText}`);
        }
        const text = await response.text();
        setContent(text);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load content');
      } finally {
        setLoading(false);
      }
    };

    loadContent();
  }, [contentPath]);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Loader2 className="w-8 h-8 animate-spin text-primary-500" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="card p-8 max-w-2xl mx-auto mt-8">
        <h2 className="text-2xl font-bold text-red-500 mb-4">Error Loading Documentation</h2>
        <p className="text-muted-foreground">{error}</p>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto">
      <div className="mb-8">
        <h1 className="text-4xl font-bold gradient-text mb-2">{title}</h1>
      </div>
      <div className="card p-8">
        <MarkdownViewer content={content} />
      </div>
    </div>
  );
}
