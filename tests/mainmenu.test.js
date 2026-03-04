import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { SHORTCUTS } from '../src/components/MainMenu.jsx';

describe('MainMenu shortcuts', () => {
  it('maps all expected shortcut keys', () => {
    assert.equal(SHORTCUTS.a, 'add');
    assert.equal(SHORTCUTS.p, 'paste');
    assert.equal(SHORTCUTS.s, 'summary');
    assert.equal(SHORTCUTS.e, 'edit');
    assert.equal(SHORTCUTS.q, 'exit');
  });

  it('has exactly 5 shortcuts', () => {
    assert.equal(Object.keys(SHORTCUTS).length, 5);
  });

  it('all shortcut values are unique', () => {
    const values = Object.values(SHORTCUTS);
    assert.equal(new Set(values).size, values.length);
  });

  it('all shortcut keys are single lowercase letters', () => {
    for (const key of Object.keys(SHORTCUTS)) {
      assert.match(key, /^[a-z]$/);
    }
  });
});
