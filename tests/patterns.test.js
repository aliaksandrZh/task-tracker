import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import {
  DATE_PATTERNS,
  TYPE_PATTERNS,
  NUMBER_PATTERNS,
  TIME_PATTERNS,
  tryPatterns,
} from '../src/parser.js';

describe('DATE_PATTERNS', () => {
  it('matches M/D/YYYY', () => {
    const r = tryPatterns(DATE_PATTERNS.map(re => ({ re, extract: (m) => m[1] })), '3/3/2026');
    assert.equal(r.extract, '3/3/2026');
  });

  it('matches MM/DD/YYYY', () => {
    const r = tryPatterns(DATE_PATTERNS.map(re => ({ re, extract: (m) => m[1] })), '12/25/2026');
    assert.equal(r.extract, '12/25/2026');
  });

  it('matches YYYY-MM-DD', () => {
    const r = tryPatterns(DATE_PATTERNS.map(re => ({ re, extract: (m) => m[1] })), '2026-03-03');
    assert.equal(r.extract, '2026-03-03');
  });

  it('does not match date embedded in text', () => {
    const r = tryPatterns(DATE_PATTERNS.map(re => ({ re, extract: (m) => m[1] })), 'foo 3/3/2026');
    assert.equal(r, null);
  });

  it('does not match incomplete date', () => {
    const r = tryPatterns(DATE_PATTERNS.map(re => ({ re, extract: (m) => m[1] })), '3/3');
    assert.equal(r, null);
  });
});

describe('TYPE_PATTERNS', () => {
  it('matches Bug', () => {
    const r = tryPatterns(TYPE_PATTERNS, 'Bug 123: name');
    assert.equal(r.extract, 'Bug');
  });

  it('matches Task', () => {
    const r = tryPatterns(TYPE_PATTERNS, 'Task 456: name');
    assert.equal(r.extract, 'Task');
  });

  it('matches case-insensitively', () => {
    assert.equal(tryPatterns(TYPE_PATTERNS, 'bug 1: x').extract, 'bug');
    assert.equal(tryPatterns(TYPE_PATTERNS, 'TASK 1: x').extract, 'TASK');
    assert.equal(tryPatterns(TYPE_PATTERNS, 'tAsK 1: x').extract, 'tAsK');
  });

  it('does not match unknown types', () => {
    assert.equal(tryPatterns(TYPE_PATTERNS, 'ask 123: name'), null);
    assert.equal(tryPatterns(TYPE_PATTERNS, 'story 123: name'), null);
  });

  it('does not match Bug/Task mid-string', () => {
    assert.equal(tryPatterns(TYPE_PATTERNS, '123 Bug name'), null);
  });
});

describe('NUMBER_PATTERNS', () => {
  it('matches plain number', () => {
    const r = tryPatterns(NUMBER_PATTERNS, '12345 some name');
    assert.equal(r.extract, '12345');
  });

  it('matches number with colon', () => {
    const r = tryPatterns(NUMBER_PATTERNS, '12345: some name');
    assert.equal(r.extract, '12345');
  });

  it('matches number with hash', () => {
    const r = tryPatterns(NUMBER_PATTERNS, '#12345: some name');
    assert.equal(r.extract, '12345');
  });

  it('matches number with hash no colon', () => {
    const r = tryPatterns(NUMBER_PATTERNS, '#999 name');
    assert.equal(r.extract, '999');
  });

  it('does not match text without leading number', () => {
    assert.equal(tryPatterns(NUMBER_PATTERNS, 'no number here'), null);
  });
});

describe('TIME_PATTERNS', () => {
  it('matches hours', () => {
    const r = tryPatterns(TIME_PATTERNS, 'some name 2h');
    assert.equal(r.extract, '2h');
  });

  it('matches minutes', () => {
    const r = tryPatterns(TIME_PATTERNS, 'some name 30m');
    assert.equal(r.extract, '30m');
  });

  it('matches compound time', () => {
    const r = tryPatterns(TIME_PATTERNS, 'some name 1h 30m');
    assert.equal(r.extract, '1h 30m');
  });

  it('matches decimal hours', () => {
    const r = tryPatterns(TIME_PATTERNS, 'some name 1.5h');
    assert.equal(r.extract, '1.5h');
  });

  it('matches decimal compound', () => {
    const r = tryPatterns(TIME_PATTERNS, 'some name 1.5h 15m');
    assert.equal(r.extract, '1.5h 15m');
  });

  it('does not match time at start of string', () => {
    assert.equal(tryPatterns(TIME_PATTERNS, '2h some name'), null);
  });

  it('does not match time embedded in name', () => {
    assert.equal(tryPatterns(TIME_PATTERNS, 'fix 2h delay issue'), null);
  });

  it('matches bare number as hours', () => {
    const r = tryPatterns(TIME_PATTERNS, 'some name 1.5');
    assert.equal(r.extract, '1.5h');
  });

  it('matches bare integer as hours', () => {
    const r = tryPatterns(TIME_PATTERNS, 'some name 2');
    assert.equal(r.extract, '2h');
  });
});
