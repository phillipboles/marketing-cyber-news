/**
 * Base API Client for NEXUS Frontend
 *
 * Typed fetch-based client with Bearer token authentication.
 * Tokens are stored in localStorage and sent via Authorization header.
 */

const API_BASE_URL =
  import.meta.env.VITE_API_URL || 'http://localhost:8080/v1';

/**
 * Token storage keys
 */
const TOKEN_STORAGE_KEY = 'aci_access_token';
const REFRESH_TOKEN_STORAGE_KEY = 'aci_refresh_token';

/**
 * Token storage utilities
 */
export const tokenStorage = {
  getAccessToken: (): string | null => localStorage.getItem(TOKEN_STORAGE_KEY),
  getRefreshToken: (): string | null => localStorage.getItem(REFRESH_TOKEN_STORAGE_KEY),
  setTokens: (accessToken: string, refreshToken: string): void => {
    localStorage.setItem(TOKEN_STORAGE_KEY, accessToken);
    localStorage.setItem(REFRESH_TOKEN_STORAGE_KEY, refreshToken);
  },
  clearTokens: (): void => {
    localStorage.removeItem(TOKEN_STORAGE_KEY);
    localStorage.removeItem(REFRESH_TOKEN_STORAGE_KEY);
  },
  hasToken: (): boolean => localStorage.getItem(TOKEN_STORAGE_KEY) !== null,
};

/**
 * API error thrown on non-2xx responses
 */
export class ApiError extends Error {
  constructor(
    message: string,
    public readonly statusCode: number,
    public readonly code: string,
    public readonly details?: Record<string, readonly string[]>
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

/**
 * API error response structure from backend
 */
interface ErrorResponse {
  readonly error: {
    readonly code: string;
    readonly message: string;
    readonly details?: Record<string, readonly string[]>;
  };
}

/**
 * Type guard to check if response is an error
 */
function isErrorResponse(data: unknown): data is ErrorResponse {
  return (
    typeof data === 'object' &&
    data !== null &&
    'error' in data &&
    typeof (data as ErrorResponse).error === 'object'
  );
}

/**
 * Base API Client class
 * Handles HTTP communication with aci-backend REST API
 */
class BaseApiClient {
  private readonly baseUrl: string;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  /**
   * Make a fetch request with standard configuration
   */
  private async request<T>(
    method: string,
    path: string,
    options?: {
      readonly body?: unknown;
      readonly params?: Record<string, string>;
    }
  ): Promise<T> {
    const url = new URL(`${this.baseUrl}${path}`);

    // Add query parameters if provided
    if (options?.params) {
      Object.entries(options.params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          url.searchParams.set(key, value);
        }
      });
    }

    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    };

    // Add Authorization header if token exists
    const accessToken = tokenStorage.getAccessToken();
    if (accessToken) {
      (headers as Record<string, string>)['Authorization'] = `Bearer ${accessToken}`;
    }

    const config: RequestInit = {
      method,
      headers,
      credentials: 'include', // Keep for CORS
    };

    // Add body for non-GET requests
    if (options?.body !== undefined) {
      config.body = JSON.stringify(options.body);
    }

    const response = await fetch(url.toString(), config);

    // Parse response body
    let data: unknown;
    try {
      data = await response.json();
    } catch {
      // Non-JSON response
      throw new ApiError(
        'Invalid response format',
        response.status,
        'INVALID_RESPONSE'
      );
    }

    // Handle non-2xx responses
    if (!response.ok) {
      if (isErrorResponse(data)) {
        throw new ApiError(
          data.error.message,
          response.status,
          data.error.code,
          data.error.details
        );
      }

      throw new ApiError(
        'Request failed',
        response.status,
        'UNKNOWN_ERROR'
      );
    }

    return data as T;
  }

  /**
   * GET request
   */
  async get<T>(
    path: string,
    params?: Record<string, string>
  ): Promise<T> {
    return this.request<T>('GET', path, { params });
  }

  /**
   * POST request
   */
  async post<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>('POST', path, { body });
  }

  /**
   * PUT request
   */
  async put<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>('PUT', path, { body });
  }

  /**
   * PATCH request
   */
  async patch<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>('PATCH', path, { body });
  }

  /**
   * DELETE request
   */
  async delete<T>(path: string): Promise<T> {
    return this.request<T>('DELETE', path);
  }
}

/**
 * Singleton API client instance
 *
 * @example
 * import { apiClient } from '@/services/api/client';
 *
 * const threats = await apiClient.get<PaginatedResponse<Threat>>('/threats', { page: '1' });
 * await apiClient.post<Bookmark>('/bookmarks', { threatId: '123' });
 */
export const apiClient = new BaseApiClient();
