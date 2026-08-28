import { SessionStorageService } from '@/storage/storageService';
import { saveProjectsParams, getProjectsParams, saveProjectTabParams, getProjectTabParams, getAllProjectTabParams, clearProjectParams } from './paramsStore';

describe('paramsStore', () => {
  beforeEach(() => {
    // clear session storage between tests
    SessionStorageService.set('__test__clear__', '1');
    sessionStorage.clear();
  });

  it('saves and retrieves projects params', () => {
    saveProjectsParams('search=foo&page=2');
    expect(getProjectsParams()).toBe('search=foo&page=2');
  });

  it('saves and retrieves per-project tab params', () => {
    saveProjectTabParams('p1', 'project', 'search=abc');
    saveProjectTabParams('p1', 'resources', 'page=3');
    expect(getProjectTabParams('p1', 'project')).toBe('search=abc');
    expect(getProjectTabParams('p1', 'resources')).toBe('page=3');
    const all = getAllProjectTabParams('p1');
    expect(all.project).toBe('search=abc');
    expect(all.resources).toBe('page=3');
    clearProjectParams('p1');
    expect(getAllProjectTabParams('p1')).toEqual({});
  });
});
