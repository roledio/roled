import type { HttpClient } from '@/services/core/httpClient';

export type Project = {
  id: string;
  created_at: string;
  updated_at: string;
  code: string;
  name: string;
  description: string;
  redirect_uris?: Array<{ redirect_uri: string; login_url?: string | null }>;
  logo_url?: string | null;
  is_active: boolean;
};

export type ProjectClient = {
  id: string;
  created_at: string;
  updated_at: string;
  code: string;
  name: string;
  project_id: string;
  is_default: boolean;
  is_active: boolean;
};

export type Resource = {
  id: string;
  name: string;
  description?: string;
  is_default?: boolean;
  permissions?: Array<{ id: string; name: string; description?: string; is_default?: boolean }>;
};

export type PermissionItem = {
  id: string;
  created_at: string;
  updated_at: string;
  name: string;
  description?: string;
  is_default?: boolean;
  resource_id: string;
  resource_name: string;
};

export type Role = {
  id: string;
  created_at: string;
  updated_at: string;
  code: string;
  name: string;
  description: string;
  is_default_signup?: boolean;
};

export type User = {
  id: string;
  created_at: string;
  updated_at: string;
  email: string;
  external_user_id?: string | null;
  display_name: string;
  is_active: boolean;
  is_email_verified: boolean;
  role_name?: string;
  avatar_url?: string | null;
};

export type PaginationInfo = {
  page_num: number;
  page_size: number;
  total_data: number;
};

export type ApiResponse<T> = {
  success: boolean;
  data: T;
  error?: { message: string };
  pagination?: PaginationInfo;
};

export type ServiceContext = {
  httpClient: HttpClient;
  baseUrl: string;
};
