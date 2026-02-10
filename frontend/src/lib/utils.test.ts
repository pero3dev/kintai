import { describe, it, expect } from 'vitest';
import { cn } from './utils';

describe('cn utility', () => {
  it('should merge class names', () => {
    expect(cn('foo', 'bar')).toBe('foo bar');
  });

  it('should handle conditional classes', () => {
    const condition = false;
    expect(cn('foo', condition && 'bar', 'baz')).toBe('foo baz');
  });

  it('should merge tailwind classes correctly', () => {
    expect(cn('px-2', 'px-4')).toBe('px-4');
  });

  it('should handle empty inputs', () => {
    expect(cn()).toBe('');
  });

  it('should handle arrays', () => {
    expect(cn(['foo', 'bar'])).toBe('foo bar');
  });

  it('should handle object syntax and falsey values', () => {
    expect(cn('foo', { bar: true, baz: false }, null, undefined, false)).toBe('foo bar');
  });

  it('should flatten nested arrays and preserve merged tailwind result', () => {
    expect(cn(['px-2', ['text-sm']], { 'px-6': true })).toBe('text-sm px-6');
  });
});

