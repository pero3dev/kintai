import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import i18n from './index';

describe('i18n setup', () => {
  beforeEach(async () => {
    await i18n.changeLanguage('ja');
  });

  afterEach(async () => {
    await i18n.changeLanguage('ja');
  });

  it('initializes with Japanese resources and fallback language', () => {
    expect(i18n.isInitialized).toBe(true);
    expect(i18n.language).toBe('ja');
    expect(i18n.languages).toContain('ja');
    expect(i18n.hasResourceBundle('ja', 'translation')).toBe(true);
    expect(i18n.hasResourceBundle('en', 'translation')).toBe(true);
  });

  it('resolves translations in both default and switched language', async () => {
    expect(i18n.t('common.login')).not.toBe('common.login');

    await i18n.changeLanguage('en');

    expect(i18n.t('common.login')).toBe('Login');
  });

  it('falls back to Japanese when a key is missing in the current language', async () => {
    i18n.addResource('ja', 'translation', 'test.onlyJa', 'ja-only-value');

    await i18n.changeLanguage('en');

    expect(i18n.t('test.onlyJa')).toBe('ja-only-value');
  });

  it('keeps interpolation values unescaped', () => {
    i18n.addResource('ja', 'translation', 'test.rawHtml', 'value: {{value}}');

    const translated = i18n.t('test.rawHtml', { value: '<strong>tag</strong>' });

    expect(translated).toContain('<strong>tag</strong>');
  });
});
