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
}

export const site: SiteConfig = {
  name: 'nginxpm CLI',
  binary: 'nginxpm',
  repo: 'piyush-gambhir/nginxpm-cli',
  tagline: 'Nginx Proxy Manager from your terminal',
  description:
    'A fast, scriptable CLI for proxy hosts, redirections, streams, certificates, access lists, users, audit logs, and settings — built for humans and coding agents.',
  badge: 'Open-source · macOS, Linux & Windows',
  installCommand:
    'curl -sSfL https://raw.githubusercontent.com/piyush-gambhir/nginxpm-cli/main/install.sh | sh',
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
