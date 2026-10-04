export const appName = 'nginxpm CLI';
export const siteUrl = 'https://projects.piyushgambhir.com/nginxpm-cli';
export const docsRoute = '/docs';
export const docsImageRoute = '/og/docs';
export const docsContentRoute = '/llms.mdx/docs';

// Agents read the llms Markdown outside the site, where root-relative links
// (which also miss the basePath) do not resolve, so make them absolute.
export function absoluteLinks(markdown: string): string {
  return markdown
    .replace(/\]\((\/[^)]*)\)/g, (_match, path: string) => `](${siteUrl}${path})`)
    .replace(/href="(\/[^"]*)"/g, (_match, path: string) => `href="${siteUrl}${path}"`);
}

// The tracked file behind a docs page: guides live in web/content/docs.
export function sourcePath(pagePath: string): string {
  return `web/content/docs/${pagePath}`;
}

export const gitConfig = {
  user: 'piyush-gambhir',
  repo: 'nginxpm-cli',
  branch: 'main',
};
