import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import * as paramsStore from '@/lib/paramsStore';

// Mock react-router-dom's useNavigate to capture navigate calls
const navigateMock = vi.fn();
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom');
    return {
        ...actual,
        useNavigate: () => navigateMock,
    };
});

function ProjectBackButton() {
    return (
        <button onClick={() => navigateMock(`/projects${paramsStore.getProjectsParams()}`)}>Back</button>
    );
}

function ClientBackButton({ pid }: { pid: string }) {
    return (
        <button onClick={() => navigateMock(`/projects/${pid}/details${paramsStore.getProjectTabParams(pid, 'project')}`)}>BackClient</button>
    );
}

function ResourceBackButton({ pid }: { pid: string }) {
    return (
        <button onClick={() => navigateMock(`/projects/${pid}/details${paramsStore.getProjectTabParams(pid, 'resources')}`)}>BackResource</button>
    );
}

describe('navigation params restoration', () => {
    beforeEach(() => {
        navigateMock.mockReset();
        sessionStorage.clear();
    });

    it('restores saved projects params when navigating back from project details', () => {
        paramsStore.saveProjectsParams('?search=hello&page_num=2');
        render(<ProjectBackButton />);
        fireEvent.click(screen.getByText('Back'));
        expect(navigateMock).toHaveBeenCalledWith('/projects?search=hello&page_num=2');
    });

    it('restores saved project tab params for client back navigation', () => {
        const pid = 'proj-123';
        paramsStore.saveProjectTabParams(pid, 'project', '?page_num=3');
        render(<ClientBackButton pid={pid} />);
        fireEvent.click(screen.getByText('BackClient'));
        expect(navigateMock).toHaveBeenCalledWith(`/projects/${pid}/details?page_num=3`);
    });

    it('restores saved project tab params for resource back navigation', () => {
        const pid = 'proj-456';
        paramsStore.saveProjectTabParams(pid, 'resources', '?search=res&tab=resources');
        render(<ResourceBackButton pid={pid} />);
        fireEvent.click(screen.getByText('BackResource'));
        expect(navigateMock).toHaveBeenCalledWith(`/projects/${pid}/details?search=res&tab=resources`);
    });
});
