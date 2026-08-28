import React from 'react';
import { createRoot } from 'react-dom/client';
import posthog from 'posthog-js';
import { PostHogErrorBoundary, PostHogProvider } from 'posthog-js/react';
import App from './App';
import './index.css';

import { telemetry } from './lib/telemetry';

telemetry.init();

createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <PostHogProvider client={posthog}>
      <PostHogErrorBoundary>
        <App />
      </PostHogErrorBoundary>
    </PostHogProvider>
  </React.StrictMode>
);
