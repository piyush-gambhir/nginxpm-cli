import {
  Bot,
  GitBranch,
  KeyRound,
  ListChecks,
  Search,
  Zap,
  type LucideIcon,
} from 'lucide-react';

export interface Feature {
  icon: LucideIcon;
  title: string;
  body: string;
}

export interface SiteStep {
  title: string;
  body: string;
  snippet?: string;
}

export interface SiteConfig {
  /** Display name, e.g. "Acme CLI" */
  name: string;
  /** The binary invoked in examples, e.g. "acme" */
  binary: string;
  /** GitHub "owner/repo" */
  repo: string;
  /** One-line hero heading */
  tagline: string;
  /** Hero sub-paragraph */
  description: string;
  /** Small pill above the heading */
  badge: string;
  /** One-line install command shown in the hero */
  installCommand: string;
  /** Feature cards */
  features: Feature[];
  /** Title above the code block */
  exampleTitle: string;
  /** Shell example rendered in the terminal card */
  example: string;
  /** Optional: tech / query languages this CLI speaks (logo strip) */
  compatible?: string[];
  /** Optional: features section heading (default: "Everything, from one binary") */
  featuresTitle?: string;
  /** Optional: features section subheading */
  featuresSubtitle?: string;
  /** Optional: CTA band body (default mentions installing the binary) */
  ctaBody?: string;
  /** Optional: per-site accent expressed as an OKLCH color */
  accent?: string;
  /** Optional: human-readable accent name */
  accentName?: string;
  /** Optional: hex twin of the accent, for surfaces without oklch() support (OG images) */
  accentHex?: string;
  /** Optional: three-step getting-started sequence */
  steps?: SiteStep[];
}

export const site: SiteConfig = {
  name: 'nginxpm CLI',
  binary: 'nginxpm',
  repo: 'piyush-gambhir/nginxpm-cli',
  tagline: 'Nginx Proxy Manager from your terminal',
  description:
    'A fast, scriptable CLI for proxy hosts, redirections, streams, certificates, access lists, users, audit logs, and settings, built for humans and coding agents.',
  badge: 'Open-source · macOS, Linux & Windows',
  accent: 'oklch(0.75 0.15 150)',
  accentName: 'green',
  accentHex: '#5dc879',
  installCommand:
    'curl -sSfL https://raw.githubusercontent.com/piyush-gambhir/nginxpm-cli/main/install.sh | sh',
  steps: [
    {
      title: 'Install',
      body: 'Install the binary, authenticate, and start querying. No runtime, no dependencies.',
      snippet:
        'curl -sSfL https://raw.githubusercontent.com/piyush-gambhir/nginxpm-cli/main/install.sh | sh',
    },
    {
      title: 'Configure without an interactive login',
      body: 'Log in with an NPM URL, email, and password, or use environment variables and named profiles for automation.',
      snippet: 'export NGINXPM_URL=https://npm.example.com',
    },
    {
      title: 'Inspect and manage the instance',
      body: 'Manage redirection hosts, TCP/UDP streams, 404 dead hosts, and access lists alongside reverse proxies.',
      snippet: 'nginxpm status -o json',
    },
  ],
  features: [
    {
      icon: Search,
      title: 'Proxy hosts',
      body: 'List, inspect, create, update, delete, enable, and disable Nginx Proxy Manager proxy hosts from the shell.',
    },
    {
      icon: KeyRound,
      title: 'Simple connection setup',
      body: 'Log in with an NPM URL, email, and password, or use environment variables and named profiles for automation.',
    },
    {
      icon: Bot,
      title: 'Agent-friendly',
      body: '-o json|yaml for machine-readable reads, --read-only safety mode, --no-input, and environment-based configuration.',
    },
    {
      icon: GitBranch,
      title: 'Every traffic route',
      body: 'Manage redirection hosts, TCP/UDP streams, 404 dead hosts, and access lists alongside reverse proxies.',
    },
    {
      icon: Zap,
      title: 'Certificate workflows',
      body: "Create, inspect, renew, and delete SSL certificates, test HTTP reachability, and discover DNS challenge providers.",
    },
    {
      icon: ListChecks,
      title: 'Complete administration',
      body: 'Manage users and permissions, inspect the audit log, update server settings, and switch between NPM instances.',
    },
  ],
  exampleTitle: 'An eight-line tour',
  example: `# Configure without an interactive login
export NGINXPM_URL=https://npm.example.com
export NGINXPM_EMAIL=admin@example.com
export NGINXPM_PASSWORD=changeme

# Inspect and manage the instance
nginxpm status -o json
nginxpm proxy list -o json`,
  compatible: [
    "Let's Encrypt",
    "Proxy hosts",
    "Streams",
    "Access lists",
    "SSL",
    "NPM API",
  ],
};

export const repositoryUrl = `https://github.com/${site.repo}`;
export const licenseUrl = `${repositoryUrl}/blob/main/LICENSE`;
export const projectDescription =
  'nginxpm CLI is an independent, unofficial open-source command-line interface for managing Nginx Proxy Manager from the terminal.';
export const structuredDataDescription = `${projectDescription} Manage proxy hosts, streams, certificates, access lists, users, audit logs, and settings with scriptable output.`;
export const siteMetadataDescription =
  'Independent, unofficial nginxpm CLI. Agent-ready and harness-agnostic, with JSON/YAML, read-only mode, and no-input flags for hosts, streams, and certificates.';
export const siteSocialDescription =
  'nginxpm lets any shell-capable coding agent or harness manage NPM hosts, streams, and certificates with JSON/YAML, read-only mode, and no-input flags.';
