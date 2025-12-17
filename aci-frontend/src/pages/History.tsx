import { useState, useEffect, useCallback } from 'react';
import type { ReactElement } from 'react';
import type { Article } from '../types';
import { userService } from '../services/userService';
import { SEVERITY_COLORS } from '../config/severity-colors';

interface HistoryEntry {
  article: Article;
  read_at: string;
  reading_time_seconds: number;
}

interface HistoryProps {
  onArticleClick: (article: Article) => void;
}

export function History({ onArticleClick }: HistoryProps): ReactElement {
  const [history, setHistory] = useState<HistoryEntry[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  const fetchHistory = useCallback(async (): Promise<void> => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await userService.getReadingHistory(currentPage, 20);
      setHistory(response.data);
      setTotalPages(response.pagination?.total_pages || 1);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch history');
    } finally {
      setIsLoading(false);
    }
  }, [currentPage]);

  useEffect(() => {
    void fetchHistory();
  }, [fetchHistory]);

  const formatDate = (dateString: string): string => {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
    });
  };

  const formatReadingTime = (seconds: number): string => {
    if (seconds < 60) return `${seconds}s`;
    const minutes = Math.floor(seconds / 60);
    return `${minutes}m`;
  };

  return (
    <div>
      <div className="mb-8">
        <h2 className="text-2xl font-bold text-white mb-2">📜 Reading History</h2>
        <p className="text-gray-400">Articles you've read recently</p>
      </div>

      {/* Loading State */}
      {isLoading && (
        <div className="flex justify-center py-12">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
        </div>
      )}

      {/* Error State */}
      {error && !isLoading && (
        <div className="bg-red-500/20 border border-red-500 text-red-400 px-4 py-3 rounded">
          {error}
        </div>
      )}

      {/* Empty State */}
      {!isLoading && !error && history.length === 0 && (
        <div className="text-center py-12 bg-gray-800 rounded-lg">
          <p className="text-4xl mb-4">📜</p>
          <p className="text-lg text-gray-300">No reading history yet</p>
          <p className="text-sm text-gray-500 mt-2">Articles you read will appear here</p>
        </div>
      )}

      {/* History List */}
      {!isLoading && !error && history.length > 0 && (
        <>
          <div className="space-y-3">
            {history.map((entry, index) => (
              <div
                key={`${entry.article.id}-${index}`}
                onClick={() => onArticleClick(entry.article)}
                className="bg-gray-800 rounded-lg p-4 hover:bg-gray-750 transition-colors cursor-pointer flex items-center gap-4"
              >
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <span
                      className={`px-2 py-0.5 text-xs font-bold text-white rounded ${SEVERITY_COLORS[entry.article.severity]}`}
                    >
                      {entry.article.severity.toUpperCase()}
                    </span>
                    <span className="px-2 py-0.5 text-xs rounded bg-blue-500/20 text-blue-400">
                      {entry.article.category}
                    </span>
                  </div>
                  <h3 className="text-white font-medium truncate">{entry.article.title}</h3>
                </div>
                <div className="text-right text-sm text-gray-400 flex-shrink-0">
                  <p>{formatDate(entry.read_at)}</p>
                  <p className="text-xs">Read for {formatReadingTime(entry.reading_time_seconds)}</p>
                </div>
              </div>
            ))}
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex justify-center items-center gap-4 mt-8">
              <button
                onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                disabled={currentPage === 1}
                className="px-4 py-2 bg-gray-700 text-white rounded disabled:opacity-50"
              >
                Previous
              </button>
              <span className="text-gray-400">
                Page {currentPage} of {totalPages}
              </span>
              <button
                onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
                disabled={currentPage === totalPages}
                className="px-4 py-2 bg-gray-700 text-white rounded disabled:opacity-50"
              >
                Next
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
