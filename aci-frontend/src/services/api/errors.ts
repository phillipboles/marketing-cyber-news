/**
 * API Error Handling Utilities
 * Provides structured error classes and type-safe error handling for API responses
 */

/**
 * Common API error codes
 * Standardized error codes for consistent error handling across the application
 */
export enum ApiErrorCode {
  UNAUTHORIZED = 'UNAUTHORIZED',
  FORBIDDEN = 'FORBIDDEN',
  NOT_FOUND = 'NOT_FOUND',
  VALIDATION_ERROR = 'VALIDATION_ERROR',
  INTERNAL_ERROR = 'INTERNAL_ERROR',
  NETWORK_ERROR = 'NETWORK_ERROR',
  TIMEOUT_ERROR = 'TIMEOUT_ERROR',
}

/**
 * Custom API Error class
 * Extends Error with additional API-specific properties
 */
export class ApiError extends Error {
  public readonly code: string;
  public readonly status: number;
  public readonly details?: Record<string, readonly string[]>;

  constructor(
    message: string,
    code: string,
    status: number,
    details?: Record<string, readonly string[]>
  ) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
    this.status = status;
    this.details = details;

    // Maintains proper stack trace for where error was thrown (only available on V8)
    // Type guard for V8-specific captureStackTrace
    const errorConstructor = Error as typeof Error & {
      captureStackTrace?: (targetObject: object, constructorOpt?: new (...args: never[]) => unknown) => void;
    };
    if (errorConstructor.captureStackTrace) {
      errorConstructor.captureStackTrace(this, ApiError);
    }
  }
}

/**
 * User-friendly error message mapping
 * Maps technical error codes to human-readable messages
 */
const ERROR_MESSAGES: Record<string, string> = {
  [ApiErrorCode.UNAUTHORIZED]: 'Your session has expired. Please sign in again.',
  [ApiErrorCode.FORBIDDEN]: 'You do not have permission to perform this action.',
  [ApiErrorCode.NOT_FOUND]: 'The requested resource was not found.',
  [ApiErrorCode.VALIDATION_ERROR]: 'Please check your input and try again.',
  [ApiErrorCode.INTERNAL_ERROR]: 'An unexpected error occurred. Please try again later.',
  [ApiErrorCode.NETWORK_ERROR]: 'Unable to connect. Please check your internet connection.',
  [ApiErrorCode.TIMEOUT_ERROR]: 'The request took too long. Please try again.',
} as const;

/**
 * Default fallback error message
 */
const DEFAULT_ERROR_MESSAGE = 'An error occurred. Please try again.';

/**
 * Maps HTTP status codes to error codes
 */
function getErrorCodeFromStatus(status: number): string {
  if (status === 401) {
    return ApiErrorCode.UNAUTHORIZED;
  }

  if (status === 403) {
    return ApiErrorCode.FORBIDDEN;
  }

  if (status === 404) {
    return ApiErrorCode.NOT_FOUND;
  }

  if (status >= 400 && status < 500) {
    return ApiErrorCode.VALIDATION_ERROR;
  }

  if (status >= 500) {
    return ApiErrorCode.INTERNAL_ERROR;
  }

  return ApiErrorCode.INTERNAL_ERROR;
}

/**
 * Parses error details from response body
 * Handles various error response formats
 */
function parseErrorDetails(body: unknown): {
  message?: string;
  code?: string;
  details?: Record<string, readonly string[]>;
} {
  if (!body || typeof body !== 'object') {
    return {};
  }

  const bodyObj = body as Record<string, unknown>;

  // Extract message
  const message =
    typeof bodyObj.message === 'string'
      ? bodyObj.message
      : typeof bodyObj.error === 'string'
      ? bodyObj.error
      : undefined;

  // Extract code
  const code = typeof bodyObj.code === 'string' ? bodyObj.code : undefined;

  // Extract validation details
  let details: Record<string, readonly string[]> | undefined;
  if (bodyObj.details && typeof bodyObj.details === 'object') {
    const detailsObj = bodyObj.details as Record<string, unknown>;
    details = {};
    for (const [key, value] of Object.entries(detailsObj)) {
      if (Array.isArray(value) && value.every((v) => typeof v === 'string')) {
        details[key] = value as readonly string[];
      }
    }
  } else if (bodyObj.errors && typeof bodyObj.errors === 'object') {
    // Alternative format: { errors: { field: ['error1', 'error2'] } }
    const errorsObj = bodyObj.errors as Record<string, unknown>;
    details = {};
    for (const [key, value] of Object.entries(errorsObj)) {
      if (Array.isArray(value) && value.every((v) => typeof v === 'string')) {
        details[key] = value as readonly string[];
      }
    }
  }

  return { message, code, details };
}

/**
 * Creates an ApiError from a fetch Response object
 *
 * @param response - The fetch Response object
 * @param body - Optional parsed response body
 * @returns ApiError instance with status, code, and details
 */
export function createApiError(response: Response, body?: unknown): ApiError {
  const { status } = response;
  const parsed = parseErrorDetails(body);

  const code = parsed.code ?? getErrorCodeFromStatus(status);
  const message = parsed.message ?? response.statusText ?? 'Request failed';

  return new ApiError(message, code, status, parsed.details);
}

/**
 * Type guard to check if error is an ApiError instance
 *
 * @param error - Unknown error object
 * @returns true if error is ApiError
 */
export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError;
}

/**
 * Type guard to check if error is a network error
 * Detects fetch errors, connection failures, and timeouts
 *
 * @param error - Unknown error object
 * @returns true if error is a network-related error
 */
export function isNetworkError(error: unknown): boolean {
  if (!error || typeof error !== 'object') {
    return false;
  }

  // Check if it's an ApiError with network-related code
  if (isApiError(error)) {
    return (
      error.code === ApiErrorCode.NETWORK_ERROR ||
      error.code === ApiErrorCode.TIMEOUT_ERROR
    );
  }

  // Check for fetch network errors
  const errorObj = error as { name?: string; message?: string };
  if (errorObj.name === 'TypeError' || errorObj.name === 'NetworkError') {
    return true;
  }

  // Check for common network error messages
  const message = errorObj.message?.toLowerCase() ?? '';
  return (
    message.includes('network') ||
    message.includes('fetch') ||
    message.includes('connection') ||
    message.includes('timeout')
  );
}

/**
 * Type guard to check if error is an unauthorized error (401)
 *
 * @param error - Unknown error object
 * @returns true if error indicates unauthorized access
 */
export function isUnauthorizedError(error: unknown): boolean {
  if (!isApiError(error)) {
    return false;
  }

  return error.code === ApiErrorCode.UNAUTHORIZED || error.status === 401;
}

/**
 * Type guard to check if error is a forbidden error (403)
 *
 * @param error - Unknown error object
 * @returns true if error indicates forbidden access
 */
export function isForbiddenError(error: unknown): boolean {
  if (!isApiError(error)) {
    return false;
  }

  return error.code === ApiErrorCode.FORBIDDEN || error.status === 403;
}

/**
 * Type guard to check if error is a not found error (404)
 *
 * @param error - Unknown error object
 * @returns true if error indicates resource not found
 */
export function isNotFoundError(error: unknown): boolean {
  if (!isApiError(error)) {
    return false;
  }

  return error.code === ApiErrorCode.NOT_FOUND || error.status === 404;
}

/**
 * Type guard to check if error is a validation error (400, 422)
 *
 * @param error - Unknown error object
 * @returns true if error indicates validation failure
 */
export function isValidationError(error: unknown): boolean {
  if (!isApiError(error)) {
    return false;
  }

  return (
    error.code === ApiErrorCode.VALIDATION_ERROR ||
    error.status === 400 ||
    error.status === 422
  );
}

/**
 * Extracts user-friendly error message from any error type
 * Returns a safe, displayable message for end users
 *
 * @param error - Unknown error object
 * @returns User-friendly error message string
 */
export function getErrorMessage(error: unknown): string {
  // Handle null/undefined
  if (!error) {
    return DEFAULT_ERROR_MESSAGE;
  }

  // Handle ApiError instances
  if (isApiError(error)) {
    // Check for mapped message
    const mappedMessage = ERROR_MESSAGES[error.code];
    if (mappedMessage) {
      return mappedMessage;
    }

    // Fall back to error's own message
    return error.message || DEFAULT_ERROR_MESSAGE;
  }

  // Handle Error instances
  if (error instanceof Error) {
    // Check if it's a network error
    if (isNetworkError(error)) {
      return ERROR_MESSAGES[ApiErrorCode.NETWORK_ERROR];
    }

    // Return error message if it exists and is safe
    if (error.message) {
      return error.message;
    }
  }

  // Handle string errors
  if (typeof error === 'string') {
    return error;
  }

  // Handle objects with message property
  if (typeof error === 'object' && 'message' in error) {
    const message = (error as { message: unknown }).message;
    if (typeof message === 'string') {
      return message;
    }
  }

  // Ultimate fallback
  return DEFAULT_ERROR_MESSAGE;
}

/**
 * Formats validation error details into a human-readable string
 *
 * @param details - Validation error details object
 * @returns Formatted error message string
 */
export function formatValidationErrors(
  details: Record<string, readonly string[]>
): string {
  const errors = Object.entries(details).map(([field, messages]) => {
    const fieldName = field.charAt(0).toUpperCase() + field.slice(1);
    return `${fieldName}: ${messages.join(', ')}`;
  });

  return errors.join('; ');
}

/**
 * Creates a network error for fetch failures
 *
 * @param originalError - The original fetch error
 * @returns ApiError instance with network error code
 */
export function createNetworkError(originalError: Error): ApiError {
  const isTimeout = originalError.message.toLowerCase().includes('timeout');
  const code = isTimeout ? ApiErrorCode.TIMEOUT_ERROR : ApiErrorCode.NETWORK_ERROR;

  return new ApiError(
    originalError.message,
    code,
    0, // No HTTP status for network errors
    undefined
  );
}
