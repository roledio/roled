// Project form validation constants
export const PROJECT_VALIDATION = {
  name: {
    maxLength: 50,
    required: true,
  },
  description: {
    maxLength: 400,
    required: false,
  },
  redirectUri: {
    maxLength: 200,
    required: true,
  },
  loginUrl: {
    maxLength: 200,
    required: false,
  },
  redirectUris: {
    minCount: 0,
    required: false,
  },
} as const;

// Client form validation constants
export const CLIENT_VALIDATION = {
  name: {
    maxLength: 50,
    required: true,
  },
  description: {
    maxLength: 200,
    required: false,
  },
} as const;

// Resource form validation constants
export const RESOURCE_VALIDATION = {
  name: {
    maxLength: 50,
    required: true,
  },
  code: {
    maxLength: 50,
    required: true,
  },
  description: {
    maxLength: 200,
    required: false,
  },
  permission: {
    name: {
      maxLength: 50,
      required: true,
    },
    code: {
      maxLength: 50,
      required: true,
    },
    description: {
      maxLength: 200,
      required: false,
    },
  },
} as const;

// Role form validation constants
export const ROLE_VALIDATION = {
  name: {
    maxLength: 50,
    required: true,
  },
  code: {
    maxLength: 50,
    required: true,
  },
  description: {
    maxLength: 200,
    required: false,
  },
} as const;

// URL/URI validation regex patterns
const URI_REGEX = /^[a-zA-Z][a-zA-Z0-9+.-]*:.+/;
const URL_REGEX = /^https?:\/.+/;

/**
 * Validates a URI string.
 * Accepts absolute URIs (with scheme like https://) or relative paths starting with /
 */
export function isValidUri(value: string): boolean {
  if (!value.trim()) return false;
  const trimmed = value.trim();
  if (trimmed.startsWith('/')) return true;
  return URI_REGEX.test(trimmed);
}

/**
 * Validates a URL string.
 * Must start with http:// or https:// (for login URLs)
 * Empty string is considered valid (for optional fields)
 */
export function isValidUrl(value: string): boolean {
  if (!value.trim()) return true; // Optional field - empty is valid
  return URL_REGEX.test(value.trim());
}

/**
 * Validates project name
 */
export function validateProjectName(name: string): string | undefined {
  if (!name.trim()) {
    return 'Name is required';
  }
  if (name.length > PROJECT_VALIDATION.name.maxLength) {
    return `Name must be ${PROJECT_VALIDATION.name.maxLength} characters or less`;
  }
  return undefined;
}

/**
 * Validates project description
 */
export function validateProjectDescription(description: string): string | undefined {
  if (description && description.length > PROJECT_VALIDATION.description.maxLength) {
    return `Description must be ${PROJECT_VALIDATION.description.maxLength} characters or less`;
  }
  return undefined;
}

/**
 * Validates a single redirect URI row
 */
export function validateRedirectUriRow(row: {
  redirect_uri: string;
  login_url?: string;
}): { redirect_uri?: string; login_url?: string } {
  const errors: { redirect_uri?: string; login_url?: string } = {};

  // Check format
  if (row.redirect_uri.trim() && !isValidUri(row.redirect_uri)) {
    errors.redirect_uri = 'Invalid URI format';
  }

  // Check length
  if (row.redirect_uri.length > PROJECT_VALIDATION.redirectUri.maxLength) {
    errors.redirect_uri = `Redirect URI must be ${PROJECT_VALIDATION.redirectUri.maxLength} characters or less`;
  }

  // Check login URL format
  if (row.login_url?.trim() && !isValidUrl(row.login_url)) {
    errors.login_url = 'Invalid URL format';
  }

  // Check login URL length
  if (row.login_url && row.login_url.length > PROJECT_VALIDATION.loginUrl.maxLength) {
    errors.login_url = `Login URL must be ${PROJECT_VALIDATION.loginUrl.maxLength} characters or less`;
  }

  return errors;
}

/**
 * Validates redirect URIs array
 */
export function validateRedirectUris(
  redirectUris: Array<{ redirect_uri: string; login_url?: string }>
): { error?: string; rowErrors: Array<{ redirect_uri?: string; login_url?: string }>; validCount: number } {
  const rowErrors = redirectUris.map(validateRedirectUriRow);

  const validUris = redirectUris.filter((r, idx) => {
    const hasValue = r.redirect_uri.trim().length > 0;
    const hasValidFormat = isValidUri(r.redirect_uri);
    const hasValidLoginUrl = !rowErrors[idx].login_url;
    return hasValue && hasValidFormat && hasValidLoginUrl;
  });

  let error: string | undefined;

  const hasAnyInput = redirectUris.some((r) => r.redirect_uri.trim() || r.login_url?.trim());

  if (hasAnyInput && validUris.length === 0) {
    const anyInvalidFormat = redirectUris.some(
      (r) => r.redirect_uri.trim() && !isValidUri(r.redirect_uri)
    );
    const anyInvalidLoginUrl = redirectUris.some(
      (r) => r.login_url?.trim() && !isValidUrl(r.login_url)
    );

    if (anyInvalidFormat) {
      error = 'Please enter valid redirect URIs';
    } else if (anyInvalidLoginUrl) {
      error = 'Please enter valid login URLs';
    }
  }

  return { error, rowErrors, validCount: validUris.length };
}

/**
 * Complete project form validation result
 */
export interface ProjectValidationResult {
  errors: {
    name?: string;
    description?: string;
    redirectUris?: string;
  };
  rowErrors: Array<{ redirect_uri?: string; login_url?: string }>;
  isValid: boolean;
  validRedirectUris: Array<{ redirect_uri: string; login_url?: string }>;
}

/**
 * Client form validation result
 */
export interface ClientValidationResult {
  errors: {
    name?: string;
    description?: string;
  };
  isValid: boolean;
}

/**
 * Validates entire project form
 */
export function validateProjectForm(
  name: string,
  description: string,
  redirectUris: Array<{ redirect_uri: string; login_url?: string }>
): ProjectValidationResult {
  const nameError = validateProjectName(name);
  const descriptionError = validateProjectDescription(description);
  const redirectValidation = validateRedirectUris(redirectUris);

  const errors = {
    name: nameError,
    description: descriptionError,
    redirectUris: redirectValidation.error,
  };

  const hasFormatErrors = redirectValidation.rowErrors.some(
    (e) => e.redirect_uri || e.login_url
  );

  const isValid = !errors.name && !errors.description && !errors.redirectUris && !hasFormatErrors;

  return {
    errors,
    rowErrors: redirectValidation.rowErrors,
    isValid,
    validRedirectUris: redirectUris.filter((r) => {
      const hasValue = r.redirect_uri.trim().length > 0;
      const hasValidFormat = isValidUri(r.redirect_uri);
      const hasValidLoginUrl = isValidUrl(r.login_url ?? '');
      return hasValue && hasValidFormat && hasValidLoginUrl;
    }),
  };
}

/**
 * Truncates string to max length
 */
export function truncateToMaxLength(value: string, maxLength: number): string {
  if (value.length <= maxLength) return value;
  return value.slice(0, maxLength);
}

/**
 * Validates client name
 */
export function validateClientName(name: string): string | undefined {
  if (!name.trim()) {
    return 'Name is required';
  }
  if (name.length > CLIENT_VALIDATION.name.maxLength) {
    return `Name must be ${CLIENT_VALIDATION.name.maxLength} characters or less`;
  }
  return undefined;
}

/**
 * Validates client description
 */
export function validateClientDescription(description: string): string | undefined {
  if (description && description.length > CLIENT_VALIDATION.description.maxLength) {
    return `Description must be ${CLIENT_VALIDATION.description.maxLength} characters or less`;
  }
  return undefined;
}

/**
 * Validates entire client form
 */
export function validateClientForm(
  name: string,
  description: string
): ClientValidationResult {
  const nameError = validateClientName(name);
  const descriptionError = validateClientDescription(description);

  const errors = {
    name: nameError,
    description: descriptionError,
  };

  const isValid = !errors.name && !errors.description;

  return { errors, isValid };
}

// Resource validation types
export interface ResourceValidationResult {
  errors: {
    name?: string;
    code?: string;
    description?: string;
  };
  isValid: boolean;
}

export interface PermissionValidationResult {
  errors: {
    name?: string;
    code?: string;
    description?: string;
  };
  isValid: boolean;
}

/**
 * Validates resource name
 */
export function validateResourceName(name: string): string | undefined {
  if (!name.trim()) {
    return 'Name is required';
  }
  if (name.length > RESOURCE_VALIDATION.name.maxLength) {
    return `Name must be ${RESOURCE_VALIDATION.name.maxLength} characters or less`;
  }
  return undefined;
}

/**
 * Validates resource code
 */
export function validateResourceCode(code: string): string | undefined {
  if (!code.trim()) {
    return 'Code is required';
  }
  if (code.length > RESOURCE_VALIDATION.code.maxLength) {
    return `Code must be ${RESOURCE_VALIDATION.code.maxLength} characters or less`;
  }
  return undefined;
}

/**
 * Validates resource description
 */
export function validateResourceDescription(description: string): string | undefined {
  if (description && description.length > RESOURCE_VALIDATION.description.maxLength) {
    return `Description must be ${RESOURCE_VALIDATION.description.maxLength} characters or less`;
  }
  return undefined;
}

/**
 * Validates entire resource form
 */
export function validateResourceForm(
  name: string,
  code: string,
  description: string
): ResourceValidationResult {
  const nameError = validateResourceName(name);
  const codeError = validateResourceCode(code);
  const descriptionError = validateResourceDescription(description);

  const errors = {
    name: nameError,
    code: codeError,
    description: descriptionError,
  };

  const isValid = !errors.name && !errors.code && !errors.description;

  return { errors, isValid };
}

/**
 * Validates permission name
 */
export function validatePermissionName(name: string): string | undefined {
  if (!name.trim()) {
    return 'Name is required';
  }
  if (name.length > RESOURCE_VALIDATION.permission.name.maxLength) {
    return `Name must be ${RESOURCE_VALIDATION.permission.name.maxLength} characters or less`;
  }
  return undefined;
}

/**
 * Validates permission code
 */
export function validatePermissionCode(
  code: string,
  existingCodes: string[] = []
): string | undefined {
  if (!code.trim()) {
    return 'Code is required';
  }
  if (code.length > RESOURCE_VALIDATION.permission.code.maxLength) {
    return `Code must be ${RESOURCE_VALIDATION.permission.code.maxLength} characters or less`;
  }
  // Check for duplicate codes (case-insensitive)
  const normalizedCode = code.toLowerCase().trim();
  if (existingCodes.some(c => c.toLowerCase().trim() === normalizedCode)) {
    return 'Code must be unique';
  }
  return undefined;
}

/**
 * Validates permission description
 */
export function validatePermissionDescription(description: string): string | undefined {
  if (description && description.length > RESOURCE_VALIDATION.permission.description.maxLength) {
    return `Description must be ${RESOURCE_VALIDATION.permission.description.maxLength} characters or less`;
  }
  return undefined;
}

/**
 * Validates permission form
 */
export function validatePermissionForm(
  name: string,
  code: string,
  description: string,
  existingCodes: string[] = []
): PermissionValidationResult {
  const nameError = validatePermissionName(name);
  const codeError = validatePermissionCode(code, existingCodes);
  const descriptionError = validatePermissionDescription(description);

  const errors = {
    name: nameError,
    code: codeError,
    description: descriptionError,
  };

  const isValid = !errors.name && !errors.code && !errors.description;

  return { errors, isValid };
}

// Role validation types
export interface RoleValidationResult {
  errors: {
    name?: string;
    code?: string;
    description?: string;
  };
  isValid: boolean;
}

/**
 * Validates role name
 */
export function validateRoleName(name: string): string | undefined {
  if (!name.trim()) {
    return 'Name is required';
  }
  if (name.length > ROLE_VALIDATION.name.maxLength) {
    return `Name must be ${ROLE_VALIDATION.name.maxLength} characters or less`;
  }
  return undefined;
}

/**
 * Validates role code
 */
export function validateRoleCode(code: string): string | undefined {
  if (!code.trim()) {
    return 'Code is required';
  }
  if (code.length > ROLE_VALIDATION.code.maxLength) {
    return `Code must be ${ROLE_VALIDATION.code.maxLength} characters or less`;
  }
  return undefined;
}

/**
 * Validates role description
 */
export function validateRoleDescription(description: string): string | undefined {
  if (description && description.length > ROLE_VALIDATION.description.maxLength) {
    return `Description must be ${ROLE_VALIDATION.description.maxLength} characters or less`;
  }
  return undefined;
}

/**
 * Validates entire role form
 */
export function validateRoleForm(
  name: string,
  code: string,
  description: string
): RoleValidationResult {
  const nameError = validateRoleName(name);
  const codeError = validateRoleCode(code);
  const descriptionError = validateRoleDescription(description);

  const errors = {
    name: nameError,
    code: codeError,
    description: descriptionError,
  };

  const isValid = !errors.name && !errors.code && !errors.description;

  return { errors, isValid };
}

// User form validation constants
export const USER_VALIDATION = {
  name: {
    maxLength: 50,
    required: true,
  },
  email: {
    maxLength: 100,
    required: true, // Required when external_user_id is empty
  },
  password: {
    minLength: 8,
    required: true, // Required when external_user_id is empty
  },
  externalUserId: {
    maxLength: 100,
    required: true, // Required when email is empty
  },
} as const;

// User validation types
export interface UserValidationResult {
  errors: {
    name?: string;
    email?: string;
    password?: string;
    externalUserId?: string;
  };
  isValid: boolean;
}

/**
 * Validates user display name
 */
export function validateUserName(name: string): string | undefined {
  if (!name.trim()) {
    return 'Name is required';
  }
  if (name.length > USER_VALIDATION.name.maxLength) {
    return `Name must be ${USER_VALIDATION.name.maxLength} characters or less`;
  }
  return undefined;
}

/**
 * Validates email
 */
export function validateUserEmail(email: string, externalUserId: string): string | undefined {
  // Email is required if externalUserId is empty
  if (!email.trim() && !externalUserId.trim()) {
    return 'Email is required when External User ID is not provided';
  }
  if (email.trim()) {
    if (email.length > USER_VALIDATION.email.maxLength) {
      return `Email must be ${USER_VALIDATION.email.maxLength} characters or less`;
    }
    // Basic email format validation
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      return 'Please enter a valid email address';
    }
  }
  return undefined;
}

/**
 * Validates password
 */
export function validateUserPassword(password: string, externalUserId: string, isUpdate = false): string | undefined {
  // Password is required if externalUserId is empty (for new users)
  // For updates, password is optional (can be blank to keep current)
  if (!externalUserId.trim() && !password.trim() && !isUpdate) {
    return 'Password is required when External User ID is not provided';
  }
  if (password.trim() && password.length < USER_VALIDATION.password.minLength) {
    return `Password must be at least ${USER_VALIDATION.password.minLength} characters`;
  }
  return undefined;
}

/**
 * Validates external user ID
 */
export function validateUserExternalId(externalUserId: string, email: string): string | undefined {
  // External User ID is required if email is empty
  if (!externalUserId.trim() && !email.trim()) {
    return 'External User ID is required when Email is not provided';
  }
  if (externalUserId.trim() && externalUserId.length > USER_VALIDATION.externalUserId.maxLength) {
    return `External User ID must be ${USER_VALIDATION.externalUserId.maxLength} characters or less`;
  }
  return undefined;
}

/**
 * Validates entire user form
 */
export function validateUserForm(
  name: string,
  email: string,
  password: string,
  externalUserId: string,
  isUpdate = false
): UserValidationResult {
  const nameError = validateUserName(name);
  const emailError = validateUserEmail(email, externalUserId);
  const passwordError = validateUserPassword(password, externalUserId, isUpdate);
  const externalIdError = validateUserExternalId(externalUserId, email);

  const errors = {
    name: nameError,
    email: emailError,
    password: passwordError,
    externalUserId: externalIdError,
  };

  const isValid = !errors.name && !errors.email && !errors.password && !errors.externalUserId;

  return { errors, isValid };
}

// Profile validation types
export interface ProfileValidationResult {
  errors: {
    name?: string;
    email?: string;
    password?: string;
  };
  isValid: boolean;
}

/**
 * Validates profile email (always required for profile, no externalUserId dependency)
 */
export function validateProfileEmail(email: string): string | undefined {
  if (!email.trim()) {
    return 'Email is required';
  }
  if (email.length > USER_VALIDATION.email.maxLength) {
    return `Email must be ${USER_VALIDATION.email.maxLength} characters or less`;
  }
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (!emailRegex.test(email)) {
    return 'Please enter a valid email address';
  }
  return undefined;
}

/**
 * Validates profile password (always optional for profile updates)
 */
export function validateProfilePassword(password: string): string | undefined {
  if (password.trim() && password.length < USER_VALIDATION.password.minLength) {
    return `Password must be at least ${USER_VALIDATION.password.minLength} characters`;
  }
  return undefined;
}

/**
 * Validates entire profile form (name, email, password — no externalUserId)
 */
export function validateProfileForm(
  name: string,
  email: string,
  password: string,
): ProfileValidationResult {
  const nameError = validateUserName(name);
  const emailError = validateProfileEmail(email);
  const passwordError = validateProfilePassword(password);

  const errors = {
    name: nameError,
    email: emailError,
    password: passwordError,
  };

  const isValid = !errors.name && !errors.email && !errors.password;

  return { errors, isValid };
}
