/**
 * Roled Design System - Shared JavaScript
 * A modern, professional JS framework following Roled brand guidelines
 * Compatible with Go templates and can be minified for production
 */

(function (global, factory) {
  typeof exports === 'object' && typeof module !== 'undefined' ? module.exports = factory() :
    typeof define === 'function' && define.amd ? define(factory) :
      (global = typeof globalThis !== 'undefined' ? globalThis : global || self, global.Roled = factory());
})(this, (function () {
  'use strict';

  /**
   * Roled Design System Namespace
   */
  const Roled = {
    version: '1.0.0',

    /**
     * Initialize all Roled components
     */
    init() {
      this.initPasswordToggle();
      this.initFormValidation();
      this.initAuthToggle();
      this.initToastNotifications();
    },

    /**
     * Password Visibility Toggle
     * Toggles password field visibility with accessible button
     */
    initPasswordToggle() {
      document.querySelectorAll('.roled-password-wrapper').forEach(wrapper => {
        const input = wrapper.querySelector('.roled-input');
        const toggle = wrapper.querySelector('.roled-password-toggle');

        if (!input || !toggle) return;

        // Check if already initialized
        if (toggle.dataset.initialized === 'true') return;
        toggle.dataset.initialized = 'true';

        toggle.addEventListener('click', (e) => {
          e.preventDefault();

          const isPassword = input.type === 'password';
          input.type = isPassword ? 'text' : 'password';

          // Toggle aria-label
          toggle.setAttribute('aria-label', isPassword ? 'Hide password' : 'Show password');

          // Toggle icon
          const eyeOpenIcon = toggle.querySelector('.icon-eye-open');
          const eyeClosedIcon = toggle.querySelector('.icon-eye-closed');

          if (eyeOpenIcon && eyeClosedIcon) {
            eyeOpenIcon.style.display = isPassword ? 'block' : 'none';
            eyeClosedIcon.style.display = isPassword ? 'none' : 'block';
          }
        });
      });
    },

    /**
     * Form Validation Utilities
     * Provides consistent form validation feedback
     */
    initFormValidation() {
      // Initialize custom validation styles
      document.querySelectorAll('form[data-validate]').forEach(form => {
        form.setAttribute('novalidate', '');

        form.addEventListener('submit', (e) => {
          if (!form.checkValidity()) {
            e.preventDefault();
            e.stopPropagation();
            this.showValidationErrors(form);
          }

          form.classList.add('was-validated');
        });

        // Real-time validation feedback
        form.querySelectorAll('.roled-input').forEach(input => {
          const confirmSelector = input.dataset.confirmMatch;

          if (confirmSelector) {
            const confirmTarget = form.querySelector(confirmSelector);
            if (confirmTarget) {
              [input, confirmTarget].forEach(field => {
                field.addEventListener('blur', () => {
                  this.validateField(input);
                  this.validateField(confirmTarget);
                });

                field.addEventListener('input', () => {
                  this.validateField(input);
                  this.validateField(confirmTarget);
                });
              });
            }
          }

          input.addEventListener('blur', () => {
            this.validateField(input);
          });

          input.addEventListener('input', () => {
            if (input.classList.contains('is-invalid') || input.dataset.confirmMatch) {
              this.validateField(input);
            }
          });
        });
      });
    },

    /**
     * Validate a single field
     */
    validateField(input) {
      const wrapper = input.closest('.roled-form-group');
      if (!wrapper) return;

      const confirmSelector = input.dataset.confirmMatch;
      if (confirmSelector) {
        const matchInput = input.closest('form')?.querySelector(confirmSelector);
        if (matchInput) {
          if (input.value && matchInput.value && input.value !== matchInput.value) {
            input.setCustomValidity(input.dataset.confirmMatchMessage || 'Values do not match');
          } else {
            input.setCustomValidity('');
          }
        }
      }

      // Remove existing error messages
      const existingErrors = wrapper.querySelectorAll('.roled-error-message');
      existingErrors.forEach(el => el.remove());

      if (input.validity.valid) {
        input.classList.remove('is-invalid');
        input.classList.add('is-valid');
      } else {
        input.classList.remove('is-valid');
        input.classList.add('is-invalid');

        // Show error message
        const errorMessage = this.getValidationMessage(input);
        const errorEl = document.createElement('span');
        errorEl.className = 'roled-error-message';
        errorEl.textContent = errorMessage;
        wrapper.appendChild(errorEl);
      }
    },

    /**
     * Get validation message for input
     */
    getValidationMessage(input) {
      if (input.validity.valueMissing) {
        return input.dataset.requiredMessage || 'This field is required';
      }
      if (input.validity.customError) {
        return input.validationMessage || input.dataset.confirmMatchMessage || 'Values do not match';
      }
      if (input.validity.typeMismatch) {
        if (input.type === 'email') {
          return input.dataset.emailMessage || 'Please enter a valid email address';
        }
        return 'Please enter a valid value';
      }
      if (input.validity.tooShort) {
        return input.dataset.minlengthMessage ||
          `Please enter at least ${input.minLength} characters`;
      }
      if (input.validity.patternMismatch) {
        return input.dataset.patternMessage || 'Please match the requested format';
      }
      return 'Invalid value';
    },

    /**
     * Show validation errors for entire form
     */
    showValidationErrors(form) {
      const invalidInputs = form.querySelectorAll(':invalid');

      invalidInputs.forEach((input, index) => {
        if (index === 0) {
          input.focus();
        }
        this.validateField(input);
      });
    },

    /**
     * Auth Form Toggle (Sign In / Sign Up)
     * Handles switching between sign in and sign up forms
     */
    initAuthToggle() {
      const toggleLink = document.querySelector('[data-auth-toggle]');
      const authForm = document.querySelector('[data-auth-form]');

      if (!toggleLink || !authForm) return;

      // Check if already initialized
      if (toggleLink.dataset.initialized === 'true') return;
      toggleLink.dataset.initialized = 'true';

      const signupField = document.querySelector('[data-signup-field]');
      const confirmPasswordGroup = document.querySelector('[data-confirm-password]');
      const forgotPasswordLink = document.querySelector('[data-forgot-password]');
      const formTitle = document.querySelector('[data-form-title]');
      const submitBtn = document.querySelector('[data-submit-btn]');
      const formSubtitle = document.querySelector('[data-form-subtitle]');
      const toggleText = document.querySelector('[data-toggle-text]');
      const isSignupInput = document.querySelector('[data-is-signup]');

      let isSignup = isSignupInput?.value === 'true';

      const updateForm = (showSignup) => {
        // Update form title
        if (formTitle) {
          formTitle.textContent = showSignup ? 'Create account' : 'Sign in';
        }

        // Update submit button
        if (submitBtn) {
          submitBtn.textContent = showSignup ? 'Create account' : 'Sign in';
        }

        // Update subtitle
        if (formSubtitle) {
          const projectName = formSubtitle.dataset.projectName || '';
          formSubtitle.innerHTML = showSignup
            ? `Sign up for <strong>${projectName}</strong>`
            : `Sign in to <strong>${projectName}</strong>`;
        }

        // Update hidden signup flag
        if (isSignupInput) {
          isSignupInput.value = showSignup ? 'true' : 'false';
        }

        // Toggle confirm password field
        if (confirmPasswordGroup) {
          confirmPasswordGroup.style.display = showSignup ? 'block' : 'none';
          const confirmInput = confirmPasswordGroup.querySelector('.roled-input');
          if (confirmInput) {
            if (showSignup) {
              confirmInput.setAttribute('required', '');
              confirmInput.dataset.required = 'true';
            } else {
              confirmInput.removeAttribute('required');
              delete confirmInput.dataset.required;
            }
          }
        }

        // Toggle forgot password link
        if (forgotPasswordLink) {
          forgotPasswordLink.style.display = showSignup ? 'none' : 'block';
        }

        // Update toggle text
        if (toggleText) {
          toggleText.innerHTML = showSignup
            ? `Already have an account? <a href="#" data-auth-toggle class="roled-link">Sign in</a>`
            : `Don't have an account? <a href="#" data-auth-toggle class="roled-link">Sign up</a>`;
        }

        // Reset validation
        authForm.classList.remove('was-validated');
        authForm.querySelectorAll('.is-invalid, .is-valid').forEach(el => {
          el.classList.remove('is-invalid', 'is-valid');
        });
        authForm.querySelectorAll('.roled-error-message').forEach(el => el.remove());

        isSignup = showSignup;
      };

      // Initial state
      updateForm(isSignup);

      // Handle toggle clicks
      document.addEventListener('click', (e) => {
        const target = e.target.closest('[data-auth-toggle]');
        if (!target) return;

        e.preventDefault();
        updateForm(!isSignup);
      });
    },

    /**
     * Toast Notification System
     * Displays dismissible toast notifications
     */
    toastQueue: [],
    toastContainer: null,

    initToastNotifications() {
      // Create container if it doesn't exist
      if (!this.toastContainer) {
        this.toastContainer = document.createElement('div');
        this.toastContainer.className = 'roled-toast-container';
        this.toastContainer.setAttribute('role', 'alert');
        this.toastContainer.setAttribute('aria-live', 'polite');
        document.body.appendChild(this.toastContainer);
      }
    },

    /**
     * Show a toast notification
     * @param {Object} options - Toast options
     * @param {string} options.title - Toast title
     * @param {string} options.message - Toast message
     * @param {string} options.type - Toast type: 'error', 'success', 'info'
     * @param {number} options.duration - Duration in ms (default: 5000, -1 for persistent)
     */
    showToast({ title = '', message = '', type = 'info', duration = 5000 }) {
      this.initToastNotifications();

      const toast = document.createElement('div');
      toast.className = `roled-toast roled-toast-${type}`;

      toast.innerHTML = `
        <div class="roled-toast-content">
          ${title ? `<h5 class="roled-toast-title">${this.escapeHtml(title)}</h5>` : ''}
          ${message ? `<p class="roled-toast-message">${this.escapeHtml(message)}</p>` : ''}
        </div>
      `;

      this.toastContainer.appendChild(toast);

      // Auto-dismiss after duration
      if (duration > 0) {
        setTimeout(() => {
          this.dismissToast(toast);
        }, duration);
      }

      return toast;
    },

    /**
     * Dismiss a toast notification
     */
    dismissToast(toast) {
      toast.style.animation = 'roled-toast-slide-out 0.3s ease forwards';
      setTimeout(() => {
        toast.remove();
      }, 300);
    },

    /**
     * Escape HTML to prevent XSS
     */
    escapeHtml(text) {
      const div = document.createElement('div');
      div.textContent = text;
      return div.innerHTML;
    },

    /**
     * Show error toast from Go template data
     */
    showErrorFromTemplate(errorMessage, errorDebug = '', duration = 0) {
      let message = errorMessage;
      if (errorDebug) {
        message += `: ${errorDebug}`;
      }

      this.showToast({
        title: 'An error occurred',
        message: message,
        type: 'error',
        duration: duration
      });
    },

    /**
     * Password strength indicator
     */
    checkPasswordStrength(password) {
      let strength = 0;

      if (password.length >= 8) strength += 1;
      if (password.length >= 12) strength += 1;
      if (/[A-Z]/.test(password)) strength += 1;
      if (/[a-z]/.test(password)) strength += 1;
      if (/[0-9]/.test(password)) strength += 1;
      if (/[^A-Za-z0-9]/.test(password)) strength += 1;

      const levels = ['very-weak', 'weak', 'fair', 'good', 'strong', 'very-strong'];
      return {
        score: strength,
        level: levels[strength] || 'very-weak'
      };
    },

    /**
     * Utility: Debounce function
     */
    debounce(func, wait) {
      let timeout;
      return function executedFunction(...args) {
        const later = () => {
          clearTimeout(timeout);
          func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
      };
    },

    /**
     * Utility: Throttle function
     */
    throttle(func, limit) {
      let inThrottle;
      return function (...args) {
        if (!inThrottle) {
          func.apply(this, args);
          inThrottle = true;
          setTimeout(() => inThrottle = false, limit);
        }
      };
    }
  };

  // Auto-initialize when DOM is ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => Roled.init());
  } else {
    Roled.init();
  }

  return Roled;
}));
