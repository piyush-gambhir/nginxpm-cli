import { source } from '@/lib/source';
import { llms } from 'fumadocs-core/source';
import { site } from '@/lib/site';
import { getOtherSuiteProjects } from '@/lib/suite';

export const revalidate = false;

export function GET() {
  const [heading, ...sections] = llms(source).index().split('\n\n');
  const preamble =
    'nginxpm CLI is agent-ready and harness-agnostic: Claude Code, OpenAI Codex, Cursor, or any agent harness that can run shell commands can manage Nginx Proxy Manager hosts, streams, and certificates through structured JSON/YAML output, read-only mode, and no-input automation flags.';
  const index = [heading, preamble, ...sections].join('\n\n');
  const related = getOtherSuiteProjects(site.repo)
    .map(({ name, href }) => `- [${name}](${href})`)
    .join('\n');

  return new Response(`${index}\n\n## Related CLI sites\n\n${related}\n`);
}
