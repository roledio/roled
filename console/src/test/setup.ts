import "@testing-library/jest-dom";
import { webcrypto } from "node:crypto";

// Polyfill crypto.subtle for jsdom environment
Object.defineProperty(globalThis, "crypto", {
  value: webcrypto,
});

Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: () => {},
    removeListener: () => {},
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => {},
  }),
});
