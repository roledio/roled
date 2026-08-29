import type { HttpClient } from '@/services/core/httpClient';
import type { ApiResponse } from './types';

/** Shape of a project's settings as returned by the API. */
export type ProjectSettings = {
  is_signup_enabled: boolean;
  default_signup_role_id: string | null;
  is_signup_verify_email: boolean;
  is_forgot_password_enabled: boolean;
  is_allow_temp_email: boolean;
};

/**
 * Fetches the settings for a given project.
 *
 * @param httpClient - Authenticated HTTP client instance.
 * @param baseUrl    - Base URL of the auth service (e.g. `http://localhost:8080`).
 * @param projectId  - The project ID whose settings to fetch.
 */
export async function fetchProjectSettings(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
): Promise<ProjectSettings> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/settings`;
  try {
    const res = await httpClient.instanceRef.get<ApiResponse<ProjectSettings>>(url, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = (res as any)?.data?.error?.message ?? 'Invalid project settings response';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg =
      err?.response?.data?.error?.message ??
      err?.response?.data?.message ??
      err?.message ??
      'Failed to fetch project settings';
    throw new Error(msg);
  }
}

/**
 * Updates the settings for a given project.
 *
 * When `is_signup_enabled` is `false`, the API expects:
 * - `default_signup_role_id` → `null`
 * - `is_signup_verify_email` → `false`
 * - `is_allow_temp_email`    → `false`
 *
 * The caller is responsible for normalising the payload before passing it here.
 *
 * @param httpClient - Authenticated HTTP client instance.
 * @param baseUrl    - Base URL of the auth service.
 * @param projectId  - The project ID whose settings to update.
 * @param payload    - Updated settings payload.
 */
export async function updateProjectSettings(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  payload: ProjectSettings,
): Promise<ProjectSettings> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/settings`;
  try {
    const res = await httpClient.instanceRef.put<ApiResponse<ProjectSettings>>(url, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = (res as any)?.data?.error?.message ?? 'Failed to update project settings';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg =
      err?.response?.data?.error?.message ??
      err?.response?.data?.message ??
      err?.message ??
      'Failed to update project settings';
    throw new Error(msg);
  }
}
