import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { ReactNode } from 'react';
import { WikiPage } from './WikiPage';

let mockPathname = '/wiki';
let mockLanguage: 'ja' | 'en' = 'en';

vi.mock('@tanstack/react-router', () => ({
  Link: ({
    to,
    children,
    ...props
  }: {
    to: string;
    children: ReactNode;
    [key: string]: unknown;
  }) => (
    <a href={to} data-to={to} {...props}>
      {children}
    </a>
  ),
  useLocation: () => ({ pathname: mockPathname }),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    i18n: { language: mockLanguage },
  }),
}));

describe('WikiPage', () => {
  it('renders overview by default on /wiki', () => {
    mockPathname = '/wiki';
    mockLanguage = 'en';
    render(<WikiPage />);

    expect(screen.getByText('Internal Wiki')).toBeInTheDocument();
    expect(screen.getAllByText('Overview').length).toBeGreaterThan(0);
    expect(screen.getByText('Engineering Docs')).toBeInTheDocument();
  });

  it('renders section specific content on /wiki/backend', () => {
    mockPathname = '/wiki/backend';
    mockLanguage = 'en';
    render(<WikiPage />);

    expect(screen.getAllByText('Backend').length).toBeGreaterThan(0);
    expect(
      screen
        .getAllByRole('link')
        .some((link) => link.getAttribute('data-to') === '/wiki/backend' && link.className.includes('ring-1')),
    ).toBe(true);
  });

  it('falls back to overview when section slug is unknown', () => {
    mockPathname = '/wiki/not-found';
    mockLanguage = 'en';
    render(<WikiPage />);

    expect(screen.getAllByText('Overview').length).toBeGreaterThan(0);
    expect(screen.queryByText('not-found')).not.toBeInTheDocument();
  });

  it('renders Japanese wiki labels when locale is ja', () => {
    mockPathname = '/wiki';
    mockLanguage = 'ja';
    render(<WikiPage />);

    expect(screen.getByText('社内Wiki')).toBeInTheDocument();
    expect(screen.getAllByText('概要').length).toBeGreaterThan(0);
    expect(screen.getByText('技術ドキュメント')).toBeInTheDocument();
  });
});
