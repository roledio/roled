import { SessionStorageService } from '@/storage/storageService';

const PROJECTS_KEY = 'params:projects';
const PROJECT_TABS_KEY = (projectId: string) => `params:project:${projectId}:tabs`;

export function saveProjectsParams(params: string) {
  SessionStorageService.set(PROJECTS_KEY, params ?? '');
}

export function getProjectsParams(): string {
  return (SessionStorageService.get<string>(PROJECTS_KEY) ?? '');
}

export function saveProjectTabParams(projectId: string, tab: string, params: string) {
  const key = PROJECT_TABS_KEY(projectId);
  const existing = SessionStorageService.get<Record<string, string>>(key) ?? {};
  existing[tab] = params ?? '';
  SessionStorageService.set(key, existing);
}

export function getProjectTabParams(projectId: string, tab: string): string {
  const key = PROJECT_TABS_KEY(projectId);
  const existing = SessionStorageService.get<Record<string, string>>(key) ?? {};
  return existing[tab] ?? '';
}

export function getAllProjectTabParams(projectId: string): Record<string, string> {
  const key = PROJECT_TABS_KEY(projectId);
  return SessionStorageService.get<Record<string, string>>(key) ?? {};
}

export function clearProjectParams(projectId: string) {
  const key = PROJECT_TABS_KEY(projectId);
  SessionStorageService.remove(key);
}
