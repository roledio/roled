import { describe, it, expect, vi, beforeEach } from 'vitest';
import posthog from 'posthog-js';
import { telemetry } from './telemetry';

vi.mock('posthog-js', () => ({
  default: {
    init: vi.fn(),
    capture: vi.fn(),
    identify: vi.fn(),
    reset: vi.fn(),
  },
}));

describe('Telemetry Service', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('returns false if token or host is missing', () => {
    const result = telemetry.init({ projectToken: '', host: '', isDev: true });
    expect(result).toBe(false);
    expect(posthog.init).not.toHaveBeenCalled();
  });

  it('initializes posthog when token and host are present', () => {
    const result = telemetry.init({
      projectToken: 'phc_test_token',
      host: 'https://eu.i.posthog.com',
      isDev: false,
    });
    expect(result).toBe(true);
    expect(posthog.init).toHaveBeenCalledWith(
      'phc_test_token',
      expect.objectContaining({
        api_host: 'https://eu.i.posthog.com',
        person_profiles: 'identified_only',
        capture_pageview: true,
      })
    );
  });

  it('calls posthog.capture when trackEvent is invoked', () => {
    telemetry.init({
      projectToken: 'phc_test_token',
      host: 'https://eu.i.posthog.com',
      isDev: false,
    });

    telemetry.trackEvent('test_event', { key: 'value' });
    expect(posthog.capture).toHaveBeenCalledWith('test_event', { key: 'value' });
  });

  it('calls posthog.identify when identifyUser is invoked', () => {
    telemetry.init({
      projectToken: 'phc_test_token',
      host: 'https://eu.i.posthog.com',
      isDev: false,
    });

    telemetry.identifyUser('user-123', { email: 'test@example.com' });
    expect(posthog.identify).toHaveBeenCalledWith('user-123', { email: 'test@example.com' });
  });

  it('calls posthog.reset when resetUser is invoked', () => {
    telemetry.init({
      projectToken: 'phc_test_token',
      host: 'https://eu.i.posthog.com',
      isDev: false,
    });

    telemetry.resetUser();
    expect(posthog.reset).toHaveBeenCalled();
  });
});
