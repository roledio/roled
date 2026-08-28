import type { HttpClient } from '@/services/core/httpClient';
import type { ApiResponse, PaginationInfo, Project } from './types';

export type { Project };

export async function fetchProjects(
  httpClient: HttpClient,
  baseUrl: string,
  params: Record<string, any> = {},
): Promise<{ data: Project[]; pagination?: PaginationInfo }> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects`;
  try {
    const res = await httpClient.instanceRef.get<ApiResponse<Project[]>>(url, {
      params,
      headers: { 'Content-Type': 'application/json' },
    });

    if (!res.data?.success) {
      const msg = (res as any)?.data?.error?.message ?? 'Invalid projects response';
      throw new Error(msg);
    }

    return { data: res.data.data ?? [], pagination: res.data.pagination };
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to fetch projects';
    throw new Error(msg);
  }
}

export async function fetchProjectById(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
): Promise<Project> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}`;
  try {
    const res = await httpClient.instanceRef.get<ApiResponse<Project>>(url, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = (res as any)?.data?.error?.message ?? 'Invalid project response';
      throw new Error(msg);
    }
    return res.data.data as Project;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to fetch project';
    throw new Error(msg);
  }
}

export async function updateProject(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  payload: Partial<Project>,
): Promise<Project> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}`;
  try {
    const res = await httpClient.instanceRef.put<ApiResponse<Project>>(url, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = (res as any)?.data?.error?.message ?? 'Failed to update project';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to update project';
    throw new Error(msg);
  }
}

export async function deleteProject(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  payload: { name: string },
): Promise<void> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/delete`;
  try {
    const res = await httpClient.instanceRef.post<ApiResponse<void>>(url, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to delete project';
      throw new Error(msg);
    }
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to delete project';
    throw new Error(msg);
  }
}

export async function uploadProjectLogo(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  file: File,
): Promise<{ logo_url: string }> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/uploads`;
  const fd = new FormData();
  fd.append('file', file);
  fd.append('type', 'project-logo');
  try {
    const res = await httpClient.instanceRef.post<ApiResponse<{ url: string }>>(url, fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to upload logo';
      throw new Error(msg);
    }
    return { logo_url: res.data.data.url };
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to upload logo';
    throw new Error(msg);
  }
}

export async function uploadUserAvatar(
  httpClient: HttpClient,
  baseUrl: string,
  file: File,
): Promise<{ url: string }> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/uploads`;
  const fd = new FormData();
  fd.append('file', file);
  fd.append('type', 'user-avatar');
  try {
    const res = await httpClient.instanceRef.post<ApiResponse<{ url: string }>>(url, fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to upload avatar';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to upload avatar';
    throw new Error(msg);
  }
}

export async function createProject(
  httpClient: HttpClient,
  baseUrl: string,
  payload: {
    name: string;
    description?: string | null;
    logo_url?: string | null;
    redirect_uris?: Array<{ redirect_uri: string; login_url?: string }>;
  },
): Promise<Project> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects`;
  try {
    const res = await httpClient.instanceRef.post<ApiResponse<Project>>(url, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to create project';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to create project';
    throw new Error(msg);
  }
}
