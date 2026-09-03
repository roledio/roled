[![GitHub Release](https://img.shields.io/github/v/release/roledio/roled)](https://github.com/roledio/roled/releases)
[![CI](https://github.com/roledio/roled/actions/workflows/ci.yml/badge.svg)](https://github.com/roledio/roled/actions/workflows/ci.yml)
[![Auth Coverage](https://github.com/roledio/roled/raw/main/.badges/coverage-auth.svg)](https://github.com/roledio/roled/actions/workflows/build.yml)
[![Console Coverage](https://github.com/roledio/roled/raw/main/.badges/coverage-console.svg)](https://github.com/roledio/roled/actions/workflows/build.yml)
[![Quality gate status](https://sonarcloud.io/api/project_badges/measure?project=roledio_roled&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=roledio_roled)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=roledio_roled&metric=reliability_rating)](https://sonarcloud.io/summary/new_code?id=roledio_roled)
[![License](https://img.shields.io/github/license/roledio/roled)](https://github.com/roledio/roled/blob/main/LICENSE)

<p align="center">
  <img src="console/public/roled-logo-text-dark.png" alt="Roled Logo" style="height: 48px;" />
</p>

<p align="center">
  <strong>Centralized User & Role Management Platform</strong><br>
  <em>Build once, use it for all your projects. Stop building auth boilerplate and start building real products.</em>
</p>

<p align="center">
  <a href="https://roled.io">Cloud (Free Early Access)</a> •
  <a href="https://docs.roled.io">Documentation</a> •
  <a href="#quick-start">Quick Start</a>
</p>


## About Roled

User and role management is an inevitable infrastructure requirement for virtually every application. While essential, re-architecting user schemas, access control hierarchies, password reset flows, and admin interfaces for every new repository is repetitive and pulls focus away from your core product.

**Roled** eliminates this repetitive burden forever. It serves as a self-hosted or cloud-managed shared service where you configure roles, permissions, and users once—allowing all your applications to integrate seamlessly through clean REST APIs and built-in OAuth2 authentication flows.

As an open-source platform, Roled gives you complete control over your data and deployment. Host it within your private infrastructure, keep full ownership of your user data, and standardize access control across all your apps without vendor lock-in.

> [!NOTE]
> Roled is **not** an identity provider (IdP) or Single Sign-On (SSO) suite designed for cross-app global identities. Instead, it is purpose-built for teams managing **multiple projects with isolated user bases**. Users registered in one project cannot access another project unless explicitly granted access. Roled provides built-in OAuth2 authentication to power project logins, while keeping its core focus on effortless user and role administration.


## Tech Stack

| Layer | Technologies |
|---|---|
| **Backend** ([auth](./auth/)) | [Go](https://go.dev/), [Fiber](https://gofiber.io/), [MariaDB](https://mariadb.org/), [Redis](https://redis.io/) |
| **Frontend** ([console](./console/)) | [React](https://react.dev/), [Vite](https://vite.dev/), [TailwindCSS](https://tailwindcss.com/) |


## Quick Start

You can spin up a fully working Roled instance (Auth service, Admin Console, MariaDB, and Redis) locally using Docker Compose.

### 1. Configure Environment Variables

Copy the sample environment file:

```bash
cp .env.example .env
```

Open `.env` in your editor and configure the required settings:

#### Critical Port & URL Alignment:
Ensure the following ports match across services:
- `AUTH_PORT`, `BASE_URL`, and `VITE_AUTH_BASE_URL` **must use the same port** (e.g., `8080`).
- `CONSOLE_PORT` and `CONSOLE_BASE_URL` **must use the same port** (e.g., `4000`).

#### Credentials & Security Keys:
Update the default secrets and connection credentials before starting:
- `DB_PASSWORD` and `DB_ROOT_PASSWORD`: Secure MariaDB passwords.
- `REDIS_PASSWORD`: Secure Redis password.
- `ENCRYPTION_MASTER_KEY`: 32-byte secret key (for AES-256 encryption).
- `JWT_SIGNING_KEY`: 32–64 byte secret key (for HS256 JWT tokens).
- `EMAIL_SMTP_*`: Your SMTP server credentials for transactional emails (verification & password resets).


### 2. Start Services

Launch the stack and follow the logs:

```bash
docker compose up -d --build && docker compose logs -f
```


### 3. Retrieve Initial Seed Credentials

During the very first startup, Roled automatically seeds the database with initial credentials for the Admin Console and displays them in the container logs:

```text
  Welcome to Roled          
  Centralized User & Role Management Platform          
          
================================================================================          
                        INITIAL SEED CREDENTIALS                                         
================================================================================          
          
  Main Client for Roled Console          
  Client ID        : 2JW47EP4KpzREVncZDXPgn          
  Client Secret    : sUZp7nYfGRAaSwAMbj9oREzPLh674h6jmHkQRJ8CEKTgkAeXzAZrfKwtX44zH3KU          
  Redirect URI     : http://localhost:4000/signin/callback          
  Login URL        : http://localhost:4000/signin          
          
  Admin User          
  Email            : admin@roled.io          
  Password         : s2bVhBwG6ZAscjmL          
          
  Standard User          
  Email            : user@roled.io          
  Password         : MLfBz8A52f2xs9Rd          
          
================================================================================          
  Important: Please store these credentials safely for your records.          
================================================================================          
```

**Save these credentials securely**. 

You can now open your browser, visit `http://localhost:4000`, and log in to the Roled Console using the generated Admin user or Standard user credentials. The Admin user belongs to the system account and has access to the system project: Roled Console. The Standard user is a regular user with access to projects under their user account.


## Self-Hosted vs. Cloud

| Feature | Self-Hosted (Open Source) | Roled Cloud |
|---|---|---|
| **Deployment** | Your own servers / private cloud | Managed by Roled |
| **Data Ownership** | 100% full control & sovereignty | Hosted securely |
| **Maintenance** | Self-managed updates & backups | Automated updates & maintenance |
| **Pricing** | Free & Open Source | **Free** (Open for Early Adopters) |
| **Get Started** | Follow [Quick Start](#quick-start) | Sign up at [roled.io](https://roled.io) |

If you want to integrate immediately without setting up infrastructure, try **[Roled Cloud](https://roled.io)** for free.


## Contributing

Roled is open source and contributions are welcome! The project is under active development with many exciting features ahead.

### Upcoming Features

- **Official SDKs**: Faster integration with Roled for various programming languages.
- **Social Logins**: Google, Microsoft, GitHub, Facebook, and more.
- **Audit Logs**: Track and monitor all your projects' user activities.
- **Login Layouts**: Ready-to-use login page templates for quick project integration.
- **User Groups**: Define reusable sets of default users for quick assignment to new projects.
- **Users View**: Inverse management view, assign a user to multiple projects from a central user list without navigating project-by-project.
- And many more!

