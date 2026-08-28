import posthog from 'posthog-js';

export interface TelemetryConfig {
  projectToken?: string;
  host?: string;
  isDev?: boolean;
  enableInDev?: boolean;
}

let isInitialized = false;
let isDevEnv = false;
let isOptedOutInDev = false;

export const telemetry = {
  /**
   * Initializes PostHog telemetry based on environment configuration.
   */
  init(config?: TelemetryConfig): boolean {
    const token = config?.projectToken ?? import.meta.env.VITE_POSTHOG_PROJECT_TOKEN;
    const host = config?.host ?? import.meta.env.VITE_POSTHOG_HOST;
    const isDev = config?.isDev ?? import.meta.env.DEV;
    const enableInDev = config?.enableInDev ?? (import.meta.env.VITE_POSTHOG_ENABLE_IN_DEV === 'true');

    isDevEnv = !!isDev;
    isOptedOutInDev = isDevEnv && !enableInDev;

    if (!token || !host) {
      if (isDevEnv) {
        console.info('[Telemetry] PostHog project token or host not configured. Telemetry is disabled.');
      }
      return false;
    }

    try {
      posthog.init(token, {
        api_host: host,
        person_profiles: 'identified_only',
        capture_pageview: true,
        loaded: (ph) => {
          if (isOptedOutInDev) {
            ph.opt_out_capturing();
            console.info('[Telemetry] PostHog capturing opted out in DEV mode. Set VITE_POSTHOG_ENABLE_IN_DEV=true to enable.');
          } else {
            ph.opt_in_capturing();
          }
        },
      });

      isInitialized = true;
      return true;
    } catch (err) {
      console.warn('[Telemetry] Failed to initialize PostHog:', err);
      return false;
    }
  },

  /**
   * Captures a custom event with optional properties.
   */
  trackEvent(event: string, properties?: Record<string, unknown>): void {
    if (isDevEnv) {
      console.debug('[Telemetry Event]', event, properties);
    }

    if (!isInitialized) return;

    try {
      posthog.capture(event, properties);
    } catch (err) {
      console.warn(`[Telemetry] Error capturing event "${event}":`, err);
    }
  },

  /**
   * Identifies the current authenticated user with traits.
   */
  identifyUser(userId: string, traits?: Record<string, unknown>): void {
    if (isDevEnv) {
      console.debug('[Telemetry Identify]', userId, traits);
    }

    if (!isInitialized) return;

    try {
      posthog.identify(userId, traits);
    } catch (err) {
      console.warn(`[Telemetry] Error identifying user "${userId}":`, err);
    }
  },

  /**
   * Resets the current user session (used on logout).
   */
  resetUser(): void {
    if (isDevEnv) {
      console.debug('[Telemetry Reset]');
    }

    if (!isInitialized) return;

    try {
      posthog.reset();
    } catch (err) {
      console.warn('[Telemetry] Error resetting user session:', err);
    }
  },

  /**
   * Returns whether telemetry is initialized.
   */
  isInitialized(): boolean {
    return isInitialized;
  },
};
