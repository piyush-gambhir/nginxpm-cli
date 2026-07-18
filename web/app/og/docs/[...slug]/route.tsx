import { getPageImage, source } from '@/lib/source';
import { notFound } from 'next/navigation';
import { ImageResponse } from 'next/og';
import { site } from '@/lib/site';
import { siteUrl } from '@/lib/shared';

import { readFile } from 'node:fs/promises';
import { join } from 'node:path';

const fontBuffer = async (...fontPath: string[]) => {
  const data = await readFile(join(process.cwd(), 'node_modules', ...fontPath));
  return data.buffer.slice(data.byteOffset, data.byteOffset + data.byteLength) as ArrayBuffer;
};

export const revalidate = false;

const inter = fontBuffer('@fontsource', 'inter', 'files', 'inter-latin-400-normal.woff');
const jetbrainsMono = fontBuffer('@fontsource', 'jetbrains-mono', 'files', 'jetbrains-mono-latin-500-normal.woff');

export async function GET(_req: Request, { params }: RouteContext<'/og/docs/[...slug]'>) {
  const { slug } = await params;
  const page = source.getPage(slug.slice(0, -1));
  if (!page) notFound();

  return new ImageResponse(
    <div
      style={{
        width: '100%',
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'space-between',
        padding: '68px 72px',
        color: '#f3f4f1',
        background: '#131412',
        fontFamily: 'Inter',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
        <span style={{ color: '#55cf79', fontFamily: 'JetBrains Mono', fontSize: 30 }}>&gt;_</span>
        <span style={{ color: '#b6b8b3', fontFamily: 'JetBrains Mono', fontSize: 24 }}>{site.binary}</span>
      </div>
      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          gap: 24,
          padding: '42px 46px',
          borderRadius: 18,
          background: '#1e201b',
        }}
      >
        <div style={{ color: '#55cf79', fontFamily: 'JetBrains Mono', fontSize: 21, letterSpacing: '0.12em' }}>
          DOCUMENTATION
        </div>
        <div style={{ maxWidth: 980, fontSize: 66, lineHeight: 0.98, letterSpacing: '-0.045em' }}>
          {page.data.title}
        </div>
        <div style={{ maxWidth: 920, color: '#b6b8b3', fontSize: 27, lineHeight: 1.25 }}>
          {page.data.description}
        </div>
      </div>
      <div style={{ display: 'flex', justifyContent: 'space-between', color: '#7f827b', fontFamily: 'JetBrains Mono', fontSize: 19 }}>
        <span>{site.name}</span>
        <span>{`${siteUrl}/docs`}</span>
      </div>
    </div>,
    {
      width: 1200,
      height: 630,
      fonts: [
        { name: 'Inter', data: await inter, weight: 400 },
        { name: 'JetBrains Mono', data: await jetbrainsMono, weight: 500 },
      ],
    },
  );
}

export function generateStaticParams() {
  return source.getPages().map((page) => ({
    lang: page.locale,
    slug: getPageImage(page).segments,
  }));
}
