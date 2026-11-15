export const appName = 'nginxpm CLI';
export const siteUrl = 'https://projects.piyushgambhir.com/nginxpm-cli';
export const docsRoute = '/docs';
export const docsImageRoute = '/og/docs';
export const docsContentRoute = '/llms.mdx/docs';

export function absoluteLinks(markdown: string): string {
  return markdown
    .replace(/\]\((\/[^)]*)\)/g, (_match, path: string) => `](${siteUrl}${path})`)
    .replace(/href="(\/[^"]*)"/g, (_match, path: string) => `href="${siteUrl}${path}"`);
}

export function sourcePath(pagePath: string): string {
  return `web/content/docs/${pagePath}`;
}

export const gitConfig = {
  user: 'piyush-gambhir',
  repo: 'nginxpm-cli',
  branch: 'main',
};
