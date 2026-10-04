import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { ReactNode } from 'react';
import { ChartContainer, ChartTooltipContent, ChartLegendContent, type ChartConfig } from './chart';
vi.mock('recharts', () => ({ ResponsiveContainer: ({children}: {children:ReactNode}) => <div>{children}</div>, Tooltip: () => null, Legend: () => null }));
afterEach(cleanup);
const payload = [{dataKey:'visits',name:'visits',value:1200,color:'#123456',payload:{fill:'#123456'}}];
const config: ChartConfig = { visits: {label:'Visits',theme:{light:'#123456',dark:'#abcdef'}} };
function mount(child: React.ReactElement, settings = config) { return render(<ChartContainer id="traffic" config={settings}>{child}</ChartContainer>); }
describe('Chart presentation',()=>{
 it.each(['dot','line','dashed'] as const)('renders %s indicators with formatted values and themed colors',indicator=>{
  const {container}=mount(<ChartTooltipContent active payload={payload} label="visits" indicator={indicator} />);
  expect(screen.getAllByText('Visits').length).toBeGreaterThan(0);expect(screen.getByText((1200).toLocaleString())).toBeInTheDocument();expect(container.querySelector('style')?.textContent).toContain('--color-visits: #abcdef');expect(container.querySelector('[data-chart]')).toHaveAttribute('data-chart','chart-traffic');
 });
 it('uses custom value and label formatters',()=>{
  const formatter=vi.fn(()=> <span>Custom count</span>);const labelFormatter=vi.fn(()=> <span>Custom label</span>);
  mount(<ChartTooltipContent active payload={payload} formatter={formatter} labelFormatter={labelFormatter} />);
  expect(screen.getByText('Custom count')).toBeInTheDocument();expect(screen.getByText('Custom label')).toBeInTheDocument();expect(formatter).toHaveBeenCalledWith(1200,'visits',payload[0],0,payload[0].payload);
 });
 it.each([false,true])('hides empty or inactive tooltip, active=%s',active=>{const {container}=mount(<ChartTooltipContent active={active} payload={active?[]:payload} />);expect(container.querySelector('.shadow-xl')).toBeNull();});
 it('supports icons, hidden labels and fallback names',()=>{
  mount(<ChartTooltipContent active hideLabel payload={payload} />,{visits:{label:'Traffic',color:'red',icon:()=> <span>Traffic icon</span>}});
  expect(screen.getByText('Traffic icon')).toBeInTheDocument();expect(screen.getByText('Traffic')).toBeInTheDocument();
 });
 it('hides indicators and handles missing config',()=>{const {container}=mount(<ChartTooltipContent active hideIndicator payload={payload} />,{});expect(screen.getByText('visits')).toBeInTheDocument();expect(container.querySelector('style')).toBeNull();expect(container.querySelector('[style]')).toBeNull();});
 it.each(['top','bottom'] as const)('renders a %s legend with configured labels',verticalAlign=>{
  mount(<ChartLegendContent verticalAlign={verticalAlign} payload={[{value:'visits',dataKey:'visits',color:'red'}]} />);expect(screen.getByText('Visits')).toBeInTheDocument();
 });
 it('supports legend icons and hidden icons',()=>{const settings={visits:{label:'Visits',icon:()=> <span>Icon</span>}};const view=mount(<ChartLegendContent payload={[{value:'visits',dataKey:'visits',color:'red'}]} />,settings);expect(screen.getByText('Icon')).toBeInTheDocument();view.unmount();mount(<ChartLegendContent hideIcon payload={[{value:'visits',dataKey:'visits',color:'red'}]} />,settings);expect(screen.queryByText('Icon')).not.toBeInTheDocument();});
 it('renders no legend without payload',()=>{mount(<ChartLegendContent />);expect(screen.queryByText('Visits')).not.toBeInTheDocument();});
 it('requires a chart context',()=>{const log=vi.spyOn(console,'error').mockImplementation(()=>{});expect(()=>render(<ChartTooltipContent />)).toThrow('useChart must be used within a <ChartContainer />');log.mockRestore();});
});
