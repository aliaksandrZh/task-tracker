import { describe, it, beforeEach, afterEach } from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';

let tmpDir;
let originalCwd;

beforeEach(() => {
  tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'prefs-test-'));
  originalCwd = process.cwd();
  process.chdir(tmpDir);
});

afterEach(() => {
  process.chdir(originalCwd);
  fs.rmSync(tmpDir, { recursive: true });
});

async function freshImport() {
  const mod = await import(`../src/prefs.js?t=${Date.now()}-${Math.random()}`);
  return mod;
}

describe('prefs', () => {
  it('loadPrefs returns empty object when no file', async () => {
    const { loadPrefs } = await freshImport();
    assert.deepEqual(loadPrefs(), {});
  });

  it('savePrefs creates .prefs.json', async () => {
    const { savePrefs } = await freshImport();
    savePrefs({ sortBy: 'date' });
    assert.ok(fs.existsSync(path.join(tmpDir, '.prefs.json')));
  });

  it('savePrefs and loadPrefs round-trip', async () => {
    const { savePrefs, loadPrefs } = await freshImport();
    savePrefs({ sortBy: 'name', sortDir: 'desc' });
    const prefs = loadPrefs();
    assert.equal(prefs.sortBy, 'name');
    assert.equal(prefs.sortDir, 'desc');
  });

  it('savePrefs merges with existing prefs', async () => {
    const { savePrefs, loadPrefs } = await freshImport();
    savePrefs({ sortBy: 'date' });
    savePrefs({ sortDir: 'desc' });
    const prefs = loadPrefs();
    assert.equal(prefs.sortBy, 'date');
    assert.equal(prefs.sortDir, 'desc');
  });

  it('savePrefs overwrites existing keys', async () => {
    const { savePrefs, loadPrefs } = await freshImport();
    savePrefs({ sortBy: 'date' });
    savePrefs({ sortBy: 'name' });
    assert.equal(loadPrefs().sortBy, 'name');
  });

  it('loadPrefs returns empty object for corrupt file', async () => {
    const { loadPrefs } = await freshImport();
    fs.writeFileSync(path.join(tmpDir, '.prefs.json'), 'not json', 'utf8');
    assert.deepEqual(loadPrefs(), {});
  });
});
