import { Link, useLocation } from '@tanstack/react-router';
import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import overviewMarkdown from '../../../../docs/wiki/overview.md?raw';
import architectureMarkdown from '../../../../docs/wiki/architecture.md?raw';
import backendMarkdown from '../../../../docs/wiki/backend.md?raw';
import frontendMarkdown from '../../../../docs/wiki/frontend.md?raw';
import infrastructureMarkdown from '../../../../docs/wiki/infrastructure.md?raw';
import testingMarkdown from '../../../../docs/wiki/testing.md?raw';
import overviewMarkdownJa from '../../../../docs/wiki/ja/overview.md?raw';
import architectureMarkdownJa from '../../../../docs/wiki/ja/architecture.md?raw';
import backendMarkdownJa from '../../../../docs/wiki/ja/backend.md?raw';
import frontendMarkdownJa from '../../../../docs/wiki/ja/frontend.md?raw';
import infrastructureMarkdownJa from '../../../../docs/wiki/ja/infrastructure.md?raw';
import testingMarkdownJa from '../../../../docs/wiki/ja/testing.md?raw';

type WikiSectionId =
  | 'overview'
  | 'architecture'
  | 'backend'
  | 'frontend'
  | 'infrastructure'
  | 'testing';

type WikiLocale = 'ja' | 'en';

type MarkdownBlock =
  | { type: 'heading'; level: 1 | 2 | 3; text: string }
  | { type: 'paragraph'; text: string }
  | { type: 'list'; ordered: boolean; items: string[] }
  | { type: 'code'; language: string; content: string };

interface WikiSectionLocaleMeta {
  title: string;
  summary: string;
  sourcePath: string;
  content: string;
}

interface WikiSectionMeta {
  icon: string;
  route: string;
  locale: Record<WikiLocale, WikiSectionLocaleMeta>;
}

const sectionOrder: WikiSectionId[] = [
  'overview',
  'architecture',
  'backend',
  'frontend',
  'infrastructure',
  'testing',
];

const sections: Record<WikiSectionId, WikiSectionMeta> = {
  overview: {
    icon: 'home',
    route: '/wiki',
    locale: {
      en: {
        title: 'Overview',
        summary: 'Entry point for understanding the product and its development workflow.',
        sourcePath: 'docs/wiki/overview.md',
        content: overviewMarkdown,
      },
      ja: {
        title: '概要',
        summary: 'プロダクト全体像と開発フローを理解するための入口です。',
        sourcePath: 'docs/wiki/ja/overview.md',
        content: overviewMarkdownJa,
      },
    },
  },
  architecture: {
    icon: 'lan',
    route: '/wiki/architecture',
    locale: {
      en: {
        title: 'Architecture',
        summary: 'Domain boundaries and backend/frontend structure.',
        sourcePath: 'docs/wiki/architecture.md',
        content: architectureMarkdown,
      },
      ja: {
        title: 'アーキテクチャ',
        summary: 'ドメイン境界とバックエンド/フロントエンドの構成。',
        sourcePath: 'docs/wiki/ja/architecture.md',
        content: architectureMarkdownJa,
      },
    },
  },
  backend: {
    icon: 'dns',
    route: '/wiki/backend',
    locale: {
      en: {
        title: 'Backend',
        summary: 'Go API layering, modules, and implementation rules.',
        sourcePath: 'docs/wiki/backend.md',
        content: backendMarkdown,
      },
      ja: {
        title: 'バックエンド',
        summary: 'Go API のレイヤー構成、モジュール、実装ルール。',
        sourcePath: 'docs/wiki/ja/backend.md',
        content: backendMarkdownJa,
      },
    },
  },
  frontend: {
    icon: 'web',
    route: '/wiki/frontend',
    locale: {
      en: {
        title: 'Frontend',
        summary: 'Routing, page composition, and UI development conventions.',
        sourcePath: 'docs/wiki/frontend.md',
        content: frontendMarkdown,
      },
      ja: {
        title: 'フロントエンド',
        summary: 'ルーティング、画面構成、UI 開発ルール。',
        sourcePath: 'docs/wiki/ja/frontend.md',
        content: frontendMarkdownJa,
      },
    },
  },
  infrastructure: {
    icon: 'cloud',
    route: '/wiki/infrastructure',
    locale: {
      en: {
        title: 'Infrastructure',
        summary: 'Local environment, deployment-related settings, and operations.',
        sourcePath: 'docs/wiki/infrastructure.md',
        content: infrastructureMarkdown,
      },
      ja: {
        title: 'インフラ',
        summary: 'ローカル環境、デプロイ関連設定、運用の要点。',
        sourcePath: 'docs/wiki/ja/infrastructure.md',
        content: infrastructureMarkdownJa,
      },
    },
  },
  testing: {
    icon: 'fact_check',
    route: '/wiki/testing',
    locale: {
      en: {
        title: 'Testing',
        summary: 'Test strategy and command references by scope.',
        sourcePath: 'docs/wiki/testing.md',
        content: testingMarkdown,
      },
      ja: {
        title: 'テスト',
        summary: '変更範囲ごとのテスト戦略と実行コマンド。',
        sourcePath: 'docs/wiki/ja/testing.md',
        content: testingMarkdownJa,
      },
    },
  },
};

const commonCommands = ['make up', 'make test', 'cd backend; go test ./...', 'cd frontend; pnpm test'];

const wikiCopy = {
  en: {
    tag: 'Engineering Docs',
    pageTitle: 'Internal Wiki',
    pageDescription: 'Technical documentation for the attendance management system',
    lastUpdated: 'Last updated',
    updatedOn: 'February 10, 2026',
    docSourceTitle: 'Doc Source',
    docSourceDesc: 'This section is loaded from Markdown:',
    howToUpdateTitle: 'How To Update',
    updateSteps: [
      'Edit the corresponding `docs/wiki/*.md` file.',
      'Review the section in `/wiki` UI.',
      'Run `pnpm test` and open a PR.',
    ],
    commonCommandsTitle: 'Common Commands',
  },
  ja: {
    tag: '技術ドキュメント',
    pageTitle: '社内Wiki',
    pageDescription: '勤怠管理システムの技術ドキュメント',
    lastUpdated: '最終更新',
    updatedOn: '2026年2月10日',
    docSourceTitle: 'ドキュメントソース',
    docSourceDesc: 'このセクションは Markdown から読み込まれています:',
    howToUpdateTitle: '更新手順',
    updateSteps: [
      '対応する `docs/wiki/ja/*.md` または `docs/wiki/*.md` を編集する。',
      '`/wiki` 画面で表示を確認する。',
      '`pnpm test` を実行して PR を作成する。',
    ],
    commonCommandsTitle: 'よく使うコマンド',
  },
} as const;

function resolveWikiLocale(language: string): WikiLocale {
  return language.toLowerCase().startsWith('ja') ? 'ja' : 'en';
}

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

function resolveSection(pathname: string): WikiSectionId {
  const slug = pathname.split('/')[2];
  if (!slug) return 'overview';
  if (sectionOrder.includes(slug as WikiSectionId)) return slug as WikiSectionId;
  return 'overview';
}

function parseMarkdown(markdown: string): MarkdownBlock[] {
  const lines = markdown.replace(/\r\n/g, '\n').split('\n');
  const blocks: MarkdownBlock[] = [];

  let paragraphLines: string[] = [];
  let listItems: string[] = [];
  let listOrdered = false;
  let inCodeBlock = false;
  let codeLanguage = '';
  let codeLines: string[] = [];

  const flushParagraph = () => {
    if (paragraphLines.length === 0) return;
    blocks.push({ type: 'paragraph', text: paragraphLines.join(' ') });
    paragraphLines = [];
  };

  const flushList = () => {
    if (listItems.length === 0) return;
    blocks.push({ type: 'list', ordered: listOrdered, items: [...listItems] });
    listItems = [];
    listOrdered = false;
  };

  const flushCode = () => {
    if (codeLines.length === 0) {
      blocks.push({ type: 'code', language: codeLanguage, content: '' });
    } else {
      blocks.push({ type: 'code', language: codeLanguage, content: codeLines.join('\n') });
    }
    codeLines = [];
    codeLanguage = '';
  };

  for (const rawLine of lines) {
    const line = rawLine.trimEnd();
    const trimmed = line.trim();

    if (trimmed.startsWith('```')) {
      if (inCodeBlock) {
        flushCode();
        inCodeBlock = false;
      } else {
        flushParagraph();
        flushList();
        codeLanguage = trimmed.slice(3).trim();
        inCodeBlock = true;
      }
      continue;
    }

    if (inCodeBlock) {
      codeLines.push(line);
      continue;
    }

    if (trimmed === '') {
      flushParagraph();
      flushList();
      continue;
    }

    if (trimmed.startsWith('# ')) {
      flushParagraph();
      flushList();
      blocks.push({ type: 'heading', level: 1, text: trimmed.slice(2) });
      continue;
    }

    if (trimmed.startsWith('## ')) {
      flushParagraph();
      flushList();
      blocks.push({ type: 'heading', level: 2, text: trimmed.slice(3) });
      continue;
    }

    if (trimmed.startsWith('### ')) {
      flushParagraph();
      flushList();
      blocks.push({ type: 'heading', level: 3, text: trimmed.slice(4) });
      continue;
    }

    const orderedMatch = trimmed.match(/^\d+\.\s+(.*)$/);
    if (orderedMatch) {
      flushParagraph();
      if (listItems.length > 0 && !listOrdered) flushList();
      listOrdered = true;
      listItems.push(orderedMatch[1]);
      continue;
    }

    if (trimmed.startsWith('- ')) {
      flushParagraph();
      if (listItems.length > 0 && listOrdered) flushList();
      listOrdered = false;
      listItems.push(trimmed.slice(2));
      continue;
    }

    flushList();
    paragraphLines.push(trimmed);
  }

  flushParagraph();
  flushList();
  if (inCodeBlock) flushCode();

  return blocks;
}

function renderBlock(block: MarkdownBlock, key: string) {
  if (block.type === 'heading') {
    if (block.level === 1) {
      return (
        <h3 key={key} className="text-2xl font-bold gradient-text mt-2 mb-4">
          {block.text}
        </h3>
      );
    }
    if (block.level === 2) {
      return (
        <h4 key={key} className="text-lg font-semibold mt-6 mb-3">
          {block.text}
        </h4>
      );
    }
    return (
      <h5 key={key} className="text-base font-semibold mt-4 mb-2">
        {block.text}
      </h5>
    );
  }

  if (block.type === 'paragraph') {
    return (
      <p key={key} className="text-sm leading-7 text-muted-foreground">
        {block.text}
      </p>
    );
  }

  if (block.type === 'list') {
    const ListTag = block.ordered ? 'ol' : 'ul';
    return (
      <ListTag
        key={key}
        className={`space-y-2 text-sm text-muted-foreground pl-5 ${block.ordered ? 'list-decimal' : 'list-disc'}`}
      >
        {block.items.map((item) => (
          <li key={`${key}-${item}`}>{item}</li>
        ))}
      </ListTag>
    );
  }

  return (
    <pre key={key} className="glass-subtle rounded-xl p-4 text-xs leading-6 overflow-x-auto">
      <code>{block.content}</code>
    </pre>
  );
}

export function WikiPage() {
  const { i18n } = useTranslation();
  const location = useLocation();
  const locale = useMemo(() => resolveWikiLocale(i18n.language), [i18n.language]);
  const copy = wikiCopy[locale];
  const activeSectionId = useMemo(() => resolveSection(location.pathname), [location.pathname]);
  const activeSection = sections[activeSectionId];
  const activeSectionLocale = activeSection.locale[locale];
  const markdownBlocks = useMemo(
    () => parseMarkdown(activeSectionLocale.content),
    [activeSectionLocale.content],
  );

  return (
    <div className="space-y-6 animate-fade-in">
      <section className="glass-card rounded-2xl p-6 md:p-8 relative overflow-hidden">
        <div className="absolute inset-0 pointer-events-none">
          <div className="absolute -top-16 -right-12 size-44 rounded-full bg-amber-400/15 blur-2xl" />
          <div className="absolute -bottom-20 -left-8 size-48 rounded-full bg-indigo-400/15 blur-2xl" />
        </div>
        <div className="relative z-10 flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
          <div>
            <p className="text-xs uppercase tracking-wide text-amber-300/90">{copy.tag}</p>
            <h1 className="text-2xl md:text-3xl font-bold gradient-text mt-1">{copy.pageTitle}</h1>
            <p className="text-muted-foreground text-sm mt-1">{copy.pageDescription}</p>
          </div>
          <div className="glass-subtle rounded-xl px-4 py-2 text-sm">
            <span className="text-muted-foreground">{copy.lastUpdated}: </span>
            <span className="font-semibold">{copy.updatedOn}</span>
          </div>
        </div>
      </section>

      <section className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3">
        {sectionOrder.map((sectionId) => {
          const section = sections[sectionId];
          const sectionLocale = section.locale[locale];
          const isActive = sectionId === activeSectionId;
          return (
            <Link
              key={sectionId}
              to={section.route as '/'}
              className={`rounded-xl p-3 transition-all ${
                isActive ? 'glass-card ring-1 ring-amber-300/30' : 'glass-subtle hover:bg-white/10'
              }`}
            >
              <div className="flex items-center gap-2">
                <div
                  className={`size-8 rounded-lg flex items-center justify-center ${
                    isActive ? 'bg-amber-500/20 text-amber-300' : 'bg-white/5 text-muted-foreground'
                  }`}
                >
                  <MaterialIcon name={section.icon} className="text-lg" />
                </div>
                <span className="text-sm font-medium">{sectionLocale.title}</span>
              </div>
            </Link>
          );
        })}
      </section>

      <section className="grid grid-cols-1 xl:grid-cols-3 gap-6">
        <article className="xl:col-span-2 glass-card rounded-2xl p-6 md:p-7">
          <header className="mb-5">
            <h2 className="text-xl font-bold flex items-center gap-2">
              <MaterialIcon name={activeSection.icon} className="text-amber-300" />
              {activeSectionLocale.title}
            </h2>
            <p className="text-sm text-muted-foreground mt-2">{activeSectionLocale.summary}</p>
          </header>

          <div className="space-y-4">
            {markdownBlocks.map((block, index) => renderBlock(block, `${activeSectionId}-${index}`))}
          </div>
        </article>

        <aside className="space-y-4">
          <section className="glass-card rounded-2xl p-5">
            <h3 className="text-sm font-semibold mb-3">{copy.docSourceTitle}</h3>
            <p className="text-sm text-muted-foreground">{copy.docSourceDesc}</p>
            <p className="mt-2 text-xs font-mono text-muted-foreground">{activeSectionLocale.sourcePath}</p>
          </section>

          <section className="glass-card rounded-2xl p-5">
            <h3 className="text-sm font-semibold mb-3">{copy.howToUpdateTitle}</h3>
            <ol className="list-decimal pl-5 space-y-2 text-sm text-muted-foreground">
              {copy.updateSteps.map((step) => (
                <li key={step}>{step}</li>
              ))}
            </ol>
          </section>

          <section className="glass-card rounded-2xl p-5">
            <h3 className="text-sm font-semibold mb-3">{copy.commonCommandsTitle}</h3>
            <pre className="glass-subtle rounded-xl p-3 text-xs overflow-x-auto">
              {commonCommands.join('\n')}
            </pre>
          </section>
        </aside>
      </section>
    </div>
  );
}
