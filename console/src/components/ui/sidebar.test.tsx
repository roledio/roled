import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useIsMobile } from '@/hooks/use-mobile';
import { useSidebar } from './use-sidebar';
import * as S from './sidebar';
vi.mock('@/hooks/use-mobile', () => ({ useIsMobile: vi.fn(() => false) }));
beforeEach(() => { vi.mocked(useIsMobile).mockReturnValue(false); });
afterEach(cleanup);
function Status() { const s = useSidebar(); return <><output>{s.state}/{s.openMobile ? 'mobile open' : 'mobile closed'}</output><button onClick={() => s.setOpen(false)}>Collapse</button></>; }
function Menu() { return <><S.SidebarHeader><S.SidebarInput aria-label="Search navigation" /></S.SidebarHeader><S.SidebarSeparator /><S.SidebarContent><S.SidebarGroup><S.SidebarGroupLabel>Workspace</S.SidebarGroupLabel><S.SidebarGroupAction aria-label="Add workspace" /><S.SidebarGroupContent><S.SidebarMenu><S.SidebarMenuItem><S.SidebarMenuButton asChild isActive tooltip="Projects"><a href="/projects">Projects</a></S.SidebarMenuButton><S.SidebarMenuAction aria-label="Project actions" showOnHover /><S.SidebarMenuBadge>3</S.SidebarMenuBadge><S.SidebarMenuSub><S.SidebarMenuSubItem><S.SidebarMenuSubButton asChild size="sm" isActive><a href="/projects/new">New project</a></S.SidebarMenuSubButton></S.SidebarMenuSubItem></S.SidebarMenuSub></S.SidebarMenuItem></S.SidebarMenu></S.SidebarGroupContent></S.SidebarGroup></S.SidebarContent><S.SidebarFooter>Account</S.SidebarFooter></>; }
describe('Sidebar interaction',()=>{
 it('toggles by trigger, rail and keyboard and persists state',()=>{
  const click=vi.fn(); const {container}=render(<S.SidebarProvider><S.Sidebar collapsible="icon"><Menu /></S.Sidebar><S.SidebarTrigger onClick={click} /><S.SidebarRail /><S.SidebarInset><Status /></S.SidebarInset></S.SidebarProvider>);
  expect(screen.getByText('expanded/mobile closed')).toBeInTheDocument();expect(screen.getByRole('link',{name:'Projects'})).toHaveAttribute('data-active','true');
  fireEvent.click(screen.getAllByRole('button',{name:'Toggle Sidebar'})[0]);expect(click).toHaveBeenCalledOnce();expect(screen.getByText('collapsed/mobile closed')).toBeInTheDocument();expect(document.cookie).toContain('sidebar:state=false');
  fireEvent.keyDown(window,{key:'b',ctrlKey:true});expect(screen.getByText('expanded/mobile closed')).toBeInTheDocument();
  fireEvent.keyDown(window,{key:'x',ctrlKey:true});expect(screen.getByText('expanded/mobile closed')).toBeInTheDocument();
  fireEvent.click(container.querySelector('[data-sidebar="rail"]')!);expect(screen.getByText('collapsed/mobile closed')).toBeInTheDocument();
  fireEvent.change(screen.getByLabelText('Search navigation'),{target:{value:'Project'}});expect(screen.getByLabelText('Search navigation')).toHaveValue('Project');
 });
 it('reports changes to a controlled parent',()=>{
  const change=vi.fn();render(<S.SidebarProvider open onOpenChange={change}><Status /></S.SidebarProvider>);fireEvent.click(screen.getByRole('button',{name:'Collapse'}));expect(change).toHaveBeenCalledWith(false);expect(screen.getByText('expanded/mobile closed')).toBeInTheDocument();
 });
 it('opens a mobile sheet without changing desktop state',()=>{
  vi.mocked(useIsMobile).mockReturnValue(true);render(<S.SidebarProvider><S.Sidebar><Menu /></S.Sidebar><S.SidebarTrigger /><Status /></S.SidebarProvider>);fireEvent.click(screen.getAllByRole('button',{name:'Toggle Sidebar'})[0]);expect(screen.getByRole('dialog')).toBeInTheDocument();expect(screen.getByRole('link',{name:'Projects'})).toBeInTheDocument();fireEvent.keyDown(window,{key:'b',metaKey:true});expect(screen.queryByRole('dialog')).not.toBeInTheDocument();expect(screen.getByText('expanded/mobile closed')).toBeInTheDocument();
 });
 it.each(['floating','inset','sidebar'] as const)('renders right-side %s layout and skeleton placeholders',variant=>{
  const {container}=render(<S.SidebarProvider defaultOpen={false}><S.Sidebar side="right" variant={variant}><S.SidebarGroup><S.SidebarGroupLabel asChild><h2>Loading navigation</h2></S.SidebarGroupLabel><S.SidebarGroupAction asChild><button>Add</button></S.SidebarGroupAction><S.SidebarMenu><S.SidebarMenuItem><S.SidebarMenuSkeleton showIcon /><S.SidebarMenuSkeleton /><S.SidebarMenuButton tooltip={{children:'Details'}} size="lg">Details</S.SidebarMenuButton><S.SidebarMenuSub><S.SidebarMenuSubItem><S.SidebarMenuSubButton>Child</S.SidebarMenuSubButton></S.SidebarMenuSubItem></S.SidebarMenuSub></S.SidebarMenuItem></S.SidebarMenu></S.SidebarGroup></S.Sidebar></S.SidebarProvider>);
  expect(container.querySelector('[data-side="right"]')).toHaveAttribute('data-variant',variant);expect(screen.getByRole('heading',{name:'Loading navigation'})).toBeInTheDocument();expect(container.querySelectorAll('[data-sidebar="menu-skeleton"]')).toHaveLength(2);
 });
 it('renders a fixed sidebar without responsive collapse',()=>{render(<S.SidebarProvider><S.Sidebar collapsible="none"><Menu /></S.Sidebar></S.SidebarProvider>);expect(screen.getByRole('link',{name:'New project'})).toHaveAttribute('href','/projects/new');});
});
