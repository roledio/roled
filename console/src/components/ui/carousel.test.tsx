import { act, cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { Carousel, CarouselContent, CarouselItem, CarouselPrevious, CarouselNext } from './carousel';
const engine = vi.hoisted(() => ({ scrollPrev: vi.fn(), scrollNext: vi.fn(), canScrollPrev: vi.fn(() => false), canScrollNext: vi.fn(() => true), on: vi.fn(), off: vi.fn() }));
const embla = vi.hoisted(() => vi.fn());
vi.mock('embla-carousel-react', () => ({ default: embla }));
beforeEach(() => { vi.clearAllMocks(); engine.canScrollPrev.mockReturnValue(false); engine.canScrollNext.mockReturnValue(true); embla.mockReturnValue([vi.fn(), engine]); });
afterEach(cleanup);
describe('Carousel navigation', () => {
 it.each(['horizontal', 'vertical'] as const)('uses %s orientation, engine events and button state', orientation => {
  const setApi = vi.fn(); const view = render(<Carousel orientation={orientation} setApi={setApi}><CarouselContent><CarouselItem>First slide</CarouselItem><CarouselItem>Second slide</CarouselItem></CarouselContent><CarouselPrevious /><CarouselNext /></Carousel>);
  expect(embla).toHaveBeenCalledWith({ axis: orientation === 'horizontal' ? 'x' : 'y' }, undefined); expect(setApi).toHaveBeenCalledWith(engine);
  expect(screen.getByRole('region')).toHaveAttribute('aria-roledescription', 'carousel'); expect(screen.getAllByRole('group')).toHaveLength(2);
  expect(screen.getByRole('button', { name: 'Previous slide' })).toBeDisabled(); fireEvent.click(screen.getByRole('button', { name: 'Next slide' })); expect(engine.scrollNext).toHaveBeenCalledOnce();
  engine.canScrollPrev.mockReturnValue(true); engine.canScrollNext.mockReturnValue(false);
  const selection = engine.on.mock.calls.find(([event]) => event === 'select')?.[1] as (api: typeof engine) => void;
  act(() => selection(engine)); expect(screen.getByRole('button', { name: 'Next slide' })).toBeDisabled(); fireEvent.click(screen.getByRole('button', { name: 'Previous slide' })); expect(engine.scrollPrev).toHaveBeenCalledOnce();
  fireEvent.keyDown(screen.getByRole('region'), { key: 'ArrowLeft' }); fireEvent.keyDown(screen.getByRole('region'), { key: 'ArrowRight' }); fireEvent.keyDown(screen.getByRole('region'), { key: 'Enter' });
  expect(engine.scrollPrev).toHaveBeenCalledTimes(2); expect(engine.scrollNext).toHaveBeenCalledTimes(2);
  view.unmount(); expect(engine.off).toHaveBeenCalledWith('select', selection);
 });
 it('keeps navigation disabled until the engine is ready', () => {
  embla.mockReturnValue([vi.fn(), undefined]); render(<Carousel><CarouselContent><CarouselItem>Pending</CarouselItem></CarouselContent><CarouselPrevious /><CarouselNext /></Carousel>);
  expect(screen.getByRole('button', {name:'Next slide'})).toBeDisabled(); fireEvent.keyDown(screen.getByRole('region'), {key:'ArrowRight'}); expect(engine.scrollNext).not.toHaveBeenCalled();
 });
 it('requires a provider for carousel controls', () => { const log=vi.spyOn(console,'error').mockImplementation(()=>{}); expect(()=>render(<CarouselNext />)).toThrow('useCarousel must be used within a <Carousel />');log.mockRestore(); });
});
